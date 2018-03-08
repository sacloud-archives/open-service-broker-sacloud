package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
	"github.com/sacloud/open-service-broker-sacloud/util/random"
)

const (
	mariaDBProtocol      = "tcp"
	mariaDBMetaTableName = "open_service_broker_meta"
	mariaDBMetaTableDDL  = `CREATE TABLE %s.` + mariaDBMetaTableName + ` (
		binding_id VARCHAR(36),
		name VARCHAR(20),
		password VARCHAR(128)
	)`
)

func newMariaDBServiceHandler(operation, serviceID, planID string, rawParameter []byte) *databaseHandler {
	handler := &databaseHandler{
		operation:    operation,
		serviceID:    serviceID,
		planID:       planID,
		rawParameter: rawParameter,
		dialect:      &mariaDBHandler{},
	}

	switch operation {
	case operations.Provisioning:
		if len(rawParameter) == 0 {
			handler.paramErr = errors.New("mariaDBService parameter JSON is empty")
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

		idMap, ok := DatabaseIDMap["MariaDB"]
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
		handler.paramErr = fmt.Errorf("mariaDBService not support %q", operation)
	}

	return handler
}

type mariaDBHandler struct{}

func (f *mariaDBHandler) databaseAPI() iaas.DatabaseAPI {
	return sacloudAPI.MariaDB()
}

func (f *mariaDBHandler) buildConnInfo(host, dbName, user, password, salt string, port int) ConnectionInfo {
	return &mariaDBConnInfo{
		host:     host,
		dbName:   dbName,
		username: user,
		password: password,
		salt:     salt,
		port:     port,
	}
}

func (f *mariaDBHandler) prepareMetaTable(db *sql.DB, connInfo ConnectionInfo) error {
	exists, err := f.existsMetaTable(db, connInfo)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf(mariaDBMetaTableDDL, connInfo.UserName()))
	return err
}

func (f *mariaDBHandler) existsMetaTable(db *sql.DB, connInfo ConnectionInfo) (bool, error) {
	var res string
	query := `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = ? AND table_name = ? LIMIT 1;
		`
	err := db.QueryRow(query, connInfo.UserName(), mariaDBMetaTableName).Scan(&res)

	switch {
	case err == nil:
		return true, nil
	case err == sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

func (f *mariaDBHandler) readBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	var username, password string
	query := fmt.Sprintf(
		`SELECT name, AES_DECRYPT(password, SHA2(?,512)) FROM %s.%s WHERE binding_id = ? LIMIT 1`,
		connInfo.UserName(),
		mariaDBMetaTableName)
	rows, err := db.Query(query, connInfo.Salt(), bindingID)
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

func (f *mariaDBHandler) createBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	// create and add metadata
	username := random.String(20)
	password := random.String(30)

	_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE %s`, username))
	if err != nil {
		return nil, fmt.Errorf(`error creating user database %q: %s`, username, err)

	}

	createUserSQL := fmt.Sprintf(
		`CREATE USER '%s'@'%%' IDENTIFIED BY '%s'`,
		username, password)
	_, err = db.Exec(createUserSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating user %q: %s", username, err)
	}

	grantSQL := fmt.Sprintf(
		"GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, "+
			"INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES, "+
			"CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, "+
			"EXECUTE, REFERENCES, EVENT, "+
			"TRIGGER ON %s.* TO '%s'@'%%'",
		username, username)
	if _, err = db.Exec(grantSQL); err != nil {
		return nil, fmt.Errorf("error granting permission to %q: %s", username, err)
	}

	// insert metadata
	_, err = db.Exec(
		fmt.Sprintf("insert into %s values (?,?,AES_ENCRYPT(?, SHA2(?,512)))", mariaDBMetaTableName),
		bindingID,
		username,
		password,
		connInfo.Salt(),
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

func (f *mariaDBHandler) deleteBinding(db *sql.DB, record *databaseBindingRecord) error {

	exists, _ := f.existsUserDatabase(db, record.username)
	if exists {
		_, err := db.Exec(fmt.Sprintf(`DROP DATABASE %s`, record.username))
		if err != nil {
			return fmt.Errorf(`error deleting user database %q: %s`, record.username, err)
		}
	}

	_, err := db.Exec(
		fmt.Sprintf(`DELETE FROM %s WHERE binding_id = ?`, mariaDBMetaTableName),
		record.bindingID,
	)
	if err != nil {
		return fmt.Errorf(`error deleting user metadata record: %s`, err)
	}

	return nil
}

func (f *mariaDBHandler) existsUserDatabase(db *sql.DB, dbName string) (bool, error) {
	var res string
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name = ? LIMIT 1;`
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

type mariaDBConnInfo struct {
	username string
	password string
	host     string
	port     int
	dbName   string
	salt     string
}

func (c *mariaDBConnInfo) DriverName() string {
	return "mysql"
}

func (c *mariaDBConnInfo) UserName() string {
	return c.username
}

func (c *mariaDBConnInfo) Password() string {
	return c.password
}

func (c *mariaDBConnInfo) Host() string {
	return c.host
}

func (c *mariaDBConnInfo) Port() int {
	port := 3306
	if c.port > 0 {
		port = c.port
	}
	return port
}

func (c *mariaDBConnInfo) DBName() string {
	return c.dbName
}

func (c *mariaDBConnInfo) Salt() string {
	return c.salt
}

func (c *mariaDBConnInfo) FormatDSN() string {
	cfg := mysql.NewConfig()
	cfg.User = c.username
	cfg.Passwd = c.password
	cfg.Net = mariaDBProtocol
	cfg.Addr = fmt.Sprintf("%s:%d", c.host, c.Port())
	cfg.DBName = c.dbName
	return cfg.FormatDSN()
}
