package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"github.com/sacloud/open-service-broker-sacloud/util/random"

	_ "github.com/lib/pq" // nolint
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
)

const (
	postgreSQLMetaTableName = "open_service_broker_meta"
	postgreSQLMetaTableDDL  = `create table ` + postgreSQLMetaTableName + ` (
		binding_id varchar(36),
		name varchar(20),
		password varchar(128)
	)`
)

func newPostgreSQLServiceHandler(operation, serviceID, planID string, rawParameter []byte) *databaseHandler {
	handler := &databaseHandler{
		operation:    operation,
		serviceID:    serviceID,
		planID:       planID,
		rawParameter: rawParameter,
		dialect:      &postgreSQLHandler{},
	}

	switch operation {
	case operations.Provisioning:
		if len(rawParameter) == 0 {
			handler.paramErr = errors.New("postgreSQLService parameter JSON is empty")
			return handler
		}

		var p = params.DatabaseCreateParameter{}
		err := json.Unmarshal(rawParameter, &p)
		if err != nil {
			handler.paramErr = err
			return handler
		}

		err = p.Validate()
		if err != nil {
			handler.paramErr = err
			return handler
		}

		idMap, ok := DatabaseIDMap["postgres"]
		if ok {
			for k, v := range idMap.PlanIDMap {
				if v == planID {
					p.PlanID = k
					break
				}
			}
		}

		handler.parameter = &p
	case operations.Binding:
		// noop
	default:
		handler.paramErr = fmt.Errorf("postgreSQLService not support %q", operation)
	}

	return handler
}

type postgreSQLHandler struct{}

func (f *postgreSQLHandler) databaseAPI() iaas.DatabaseAPI {
	return sacloudAPI.PostgreSQL()
}

func (f *postgreSQLHandler) buildConnInfo(host, dbName, user, password, salt string, port int) ConnectionInfo {
	return &postgreSQLConnInfo{
		host:     host,
		dbName:   dbName,
		username: user,
		password: password,
		salt:     salt,
		port:     port,
	}
}

func (f *postgreSQLHandler) prepareMetaTable(db *sql.DB, connInfo ConnectionInfo) error {
	exists, err := f.existsMetaTable(db, connInfo)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = db.Exec(postgreSQLMetaTableDDL)
	return err
}

func (f *postgreSQLHandler) existsMetaTable(db *sql.DB, connInfo ConnectionInfo) (bool, error) {
	var res string
	query := `
			select table_name
			from information_schema.tables
			where table_catalog = $1 and table_name = $2 limit 1;
		`
	err := db.QueryRow(query, connInfo.UserName(), postgreSQLMetaTableName).Scan(&res)

	switch {
	case err == nil:
		return true, nil
	case err == sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

func (f *postgreSQLHandler) readBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	var username, password string
	query := fmt.Sprintf(`select name, password from %s where binding_id = $1 limit 1`, postgreSQLMetaTableName)
	rows, err := db.Query(query, bindingID)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		err = rows.Scan(&username, &password)
		if err != nil {
			return nil, err
		}

		return &databaseBindingRecord{
			bindingID: bindingID,
			username:  username,
			password:  password,
		}, nil
	}
	return nil, nil
}

func (f *postgreSQLHandler) createBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	// create and add metadata
	username := random.String(20)
	password := random.String(30)

	_, err := db.Exec(fmt.Sprintf("create role %q with password '%s' login", username, password))
	if err != nil {
		return nil, fmt.Errorf(`error creating user role %q: %s`, username, err)
	}

	// add Admin user to new role
	_, err = db.Exec(fmt.Sprintf("GRANT %q TO %q;", username, connInfo.UserName()))
	if err != nil {
		return nil, fmt.Errorf(`error grant to role %q: %s`, username, err)
	}

	_, err = db.Exec(fmt.Sprintf(`create database %q owner %q`, username, username))
	if err != nil {
		return nil, fmt.Errorf(`error creating user database %q: %s`, username, err)

	}

	// insert metadata
	_, err = db.Exec(
		fmt.Sprintf("insert into %s values ($1,$2,$3)", postgreSQLMetaTableName),
		bindingID,
		username,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating metadata record : %s", err)
	}

	return &databaseBindingRecord{
		bindingID: bindingID,
		username:  username,
		password:  password,
	}, nil
}

func (f *postgreSQLHandler) deleteBinding(db *sql.DB, record *databaseBindingRecord) error {

	exists, _ := f.existsUserDatabase(db, record.username)
	if exists {
		_, err := db.Exec(fmt.Sprintf(`drop database %q`, record.username))
		if err != nil {
			return fmt.Errorf(`error deleting user database %q: %s`, record.username, err)
		}
	}

	_, err := db.Exec(
		fmt.Sprintf(`delete from %s where binding_id = $1`, postgreSQLMetaTableName),
		record.bindingID,
	)
	if err != nil {
		return fmt.Errorf(`error deleting user metadata record: %s`, err)
	}

	return nil
}

func (f *postgreSQLHandler) existsUserDatabase(db *sql.DB, dbName string) (bool, error) {
	var res string
	query := `
		SELECT datname 
		FROM pg_database
		WHERE datname = $1 LIMIT 1;`
	err := db.QueryRow(query, dbName).Scan(&res)

	switch {
	case err == nil:
		return true, nil
	case err == sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

type postgreSQLConnInfo struct {
	username string
	password string
	host     string
	port     int
	dbName   string
	salt     string
}

func (c *postgreSQLConnInfo) DriverName() string {
	return "postgres"
}

func (c *postgreSQLConnInfo) UserName() string {
	return c.username
}

func (c *postgreSQLConnInfo) Password() string {
	return c.password
}

func (c *postgreSQLConnInfo) Host() string {
	return c.host
}

func (c *postgreSQLConnInfo) Port() int {
	port := 5432
	if c.port > 0 {
		port = c.port
	}
	return port
}

func (c *postgreSQLConnInfo) DBName() string {
	return c.dbName
}

func (c *postgreSQLConnInfo) Salt() string {
	return c.salt
}

func (c *postgreSQLConnInfo) FormatDSN() string {
	format := "postgres://%s:%s@%s:%d/%s?sslmode=disable"
	return fmt.Sprintf(format,
		c.username,
		c.password,
		c.host,
		c.Port(),
		c.dbName,
	)
}
