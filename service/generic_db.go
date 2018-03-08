package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"database/sql"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"github.com/sacloud/open-service-broker-sacloud/osb"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
	"github.com/sacloud/open-service-broker-sacloud/util/cmp"
)

// databaseAttrs implements InstanceState interface
type databaseAttrs struct {
	*sacloud.Database
	parameter *params.DatabaseCreateParameter
}

func (a *databaseAttrs) HasDiff() bool {
	if a.parameter == nil {
		return false
	}

	switchID, _ := strconv.ParseInt(a.Database.Remark.Switch.ID, 10, 64)
	ip := a.Database.Remark.Servers[0].(map[string]interface{})["IPAddress"].(string)
	maskLen := int32(a.Database.Remark.Network.NetworkMaskLen)
	defaultRoute := a.Database.Remark.Network.DefaultRoute

	p := a.parameter
	values := []cmp.CompareValue{
		{X: p.SwitchID, Y: switchID},
		{X: p.IPAddress, Y: ip},
		{X: p.MaskLen, Y: maskLen},
		{X: p.DefaultRoute, Y: defaultRoute},
	}

	return cmp.Equal(values...)

}

// databaseBinding implements BindingState interface
type databaseBinding struct {
	binding *osb.ServiceBinding
}

func (b *databaseBinding) HasDiff() bool {
	return false // not supported
}

func (b *databaseBinding) Binding() *osb.ServiceBinding {
	return b.binding
}

type databaseBindingRecord struct {
	bindingID string
	username  string
	password  string
}

type databaseFuncs interface {
	databaseAPI() iaas.DatabaseAPI
	buildConnInfo(host, dbName, user, password, salt string, port int) ConnectionInfo
	prepareMetaTable(db *sql.DB, connInfo ConnectionInfo) error
	existsMetaTable(db *sql.DB, connInfo ConnectionInfo) (bool, error)
	readBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error)
	createBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error)
	deleteBinding(db *sql.DB, record *databaseBindingRecord) error
}

type databaseHandler struct {
	operation    string
	serviceID    string
	planID       string
	rawParameter []byte

	parameter *params.DatabaseCreateParameter
	paramErr  error

	dialect databaseFuncs
}

func (s *databaseHandler) InstanceState(instanceID string) (InstanceState, error) {
	client := s.dialect.databaseAPI()
	db, err := client.Read(instanceID)
	if err != nil {
		if e, ok := err.(api.Error); ok {
			if e.ResponseCode() != http.StatusNotFound {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if db == nil {
		return nil, nil
	}

	return &databaseAttrs{
		Database:  db,
		parameter: s.parameter,
	}, nil
}

func (s *databaseHandler) BindingState(instanceID, bindingID string) (BindingState, error) {

	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		return nil, err
	}

	// connect db
	db, err := s.open(connInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint

	exists, err := s.dialect.existsMetaTable(db, connInfo)
	if err != nil {
		return nil, fmt.Errorf("error reading meta table: %s", err)
	}
	if !exists {
		return nil, nil
	}

	record, err := s.dialect.readBinding(db, connInfo, bindingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf(`error reading meta table: %s`, err)
	}
	if record == nil {
		return nil, nil
	}

	newConInfo := s.dialect.buildConnInfo(
		connInfo.Host(),
		record.username, // database name
		record.username,
		record.password,
		connInfo.Salt(),
		connInfo.Port(),
	)

	// return
	binding := &osb.ServiceBinding{
		Credentials: map[string]string{
			"host":        newConInfo.Host(),
			"port":        fmt.Sprintf("%d", newConInfo.Port()),
			"database":    record.username,
			"username":    record.username,
			"password":    record.password,
			"sslRequired": "false",
			"uri":         newConInfo.FormatDSN(),
		},
	}

	return &databaseBinding{binding: binding}, nil
}

func (s *databaseHandler) CreateInstance(instanceID string) error {
	_, err := s.dialect.databaseAPI().Create(instanceID, s.parameter)
	return err
}

func (s *databaseHandler) UpdateInstance(instanceID string) error {
	panic(errors.New("database services does not support update"))
}

func (s *databaseHandler) DeleteInstance(instanceID string) error {
	return s.dialect.databaseAPI().Delete(instanceID)
}

func (s *databaseHandler) CreateBinding(instanceID, bindingID string) (*osb.ServiceBinding, error) {

	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		return nil, err
	}

	// connect db
	db, err := s.open(connInfo)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint

	// check meta table exists
	err = s.dialect.prepareMetaTable(db, connInfo)
	if err != nil {
		return nil, fmt.Errorf("creating metadata table is failed: %s", err)
	}

	// check already exists
	binding, err := s.dialect.readBinding(db, connInfo, bindingID)
	if err != nil {
		return nil, fmt.Errorf("reading metadata table is failed: %s", err)
	}
	if binding != nil {
		return nil, &osb.BindingAlreadyExistsError{}
	}

	// create and add metadata
	record, err := s.dialect.createBinding(db, connInfo, bindingID)
	if err != nil {
		return nil, fmt.Errorf("creating user database is failed: %s", err)
	}
	if record == nil {
		return nil, errors.New("creating user database is failed: resulet is nil")
	}

	newConInfo := s.dialect.buildConnInfo(
		connInfo.Host(),
		record.username,
		record.username,
		record.password,
		connInfo.Salt(),
		connInfo.Port(),
	)
	// return
	return &osb.ServiceBinding{
		Credentials: map[string]string{
			"host":        newConInfo.Host(),
			"port":        fmt.Sprintf("%d", newConInfo.Port()),
			"database":    record.username,
			"username":    record.username,
			"password":    record.password,
			"sslRequired": "false",
			"uri":         newConInfo.FormatDSN(),
		},
	}, nil
}

func (s *databaseHandler) DeleteBinding(instanceID, bindingID string) error {
	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		return err
	}

	// connect db
	db, err := s.open(connInfo)
	if err != nil {
		return err
	}
	defer db.Close() // nolint

	exists, err := s.dialect.existsMetaTable(db, connInfo)
	if err != nil {
		return fmt.Errorf("error reading meta table: %s", err)
	}
	if !exists {
		return nil
	}

	// check exists meta record
	record, err := s.dialect.readBinding(db, connInfo, bindingID)
	if err != nil {
		return fmt.Errorf("reading metadata table is failed: %s", err)
	}
	if record == nil {
		return nil
	}

	// delete meta
	err = s.dialect.deleteBinding(db, record)
	if err != nil {
		return fmt.Errorf("deleting binding is failed: %s", err)
	}

	return nil
}

func (s *databaseHandler) IsValid() (bool, error) {
	return s.paramErr == nil, s.paramErr
}

func (s *databaseHandler) open(info ConnectionInfo) (*sql.DB, error) {
	if info == nil {
		return nil, errors.New("ConnectionInfo is nil")
	}

	return sql.Open(info.DriverName(), info.FormatDSN())
}

func (s *databaseHandler) connInfo(instanceID string) (ConnectionInfo, error) {

	db, err := s.dialect.databaseAPI().Read(instanceID)
	if err != nil {
		return nil, err
	}

	var port int
	if p, err := strconv.Atoi(db.Settings.DBConf.Common.ServicePort); err != nil {
		port = p
	}
	ip := db.Remark.Servers[0].(map[string]interface{})["IPAddress"].(string)

	return s.dialect.buildConnInfo(
		ip,
		db.Settings.DBConf.Common.DefaultUser,
		db.Settings.DBConf.Common.DefaultUser,
		db.Settings.DBConf.Common.UserPassword,
		db.GetStrID(),
		port,
	), nil
}
