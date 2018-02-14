package service

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/sacloud/open-service-broker-sacloud/iaas"
	"github.com/sacloud/open-service-broker-sacloud/osb"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	sql.Register("dummy", &dummyDriver{})
	sacloudAPI = testAPI
}

var testAPI = &dummyAPI{
	dbAPI: testDBAPI,
}

var testDBAPI = &genericDBDummyAPI{}

var testConnInfo = &dummyConnInfo{
	userName: "dummy",
	password: "dummy",
	host:     "192.2.0.1",
	port:     3306,
	dbName:   "dummy",
}

type genericDBDummyAPI struct {
	readResult   *sacloud.Database
	createResult *sacloud.Database
	readErr      error
	createErr    error
	deleteErr    error
}

func (c *genericDBDummyAPI) Read(instanceID string) (*sacloud.Database, error) {
	return c.readResult, c.readErr
}

func (c *genericDBDummyAPI) Create(instanceID string, param *params.DatabaseCreateParameter) (*sacloud.Database, error) {
	return c.createResult, c.createErr
}

func (c *genericDBDummyAPI) Delete(instanceID string) error {
	return c.deleteErr
}

type dummyConnInfo struct {
	userName string
	password string
	host     string
	port     int
	salt     string
	dbName   string
}

func (c *dummyConnInfo) DriverName() string {
	return "dummy"
}

func (c *dummyConnInfo) UserName() string {
	return c.userName
}

func (c *dummyConnInfo) Password() string {
	return c.password
}

func (c *dummyConnInfo) Host() string {
	return c.host
}

func (c *dummyConnInfo) Port() int {
	return c.port
}

func (c *dummyConnInfo) DBName() string {
	return c.dbName
}

func (c *dummyConnInfo) Salt() string {
	return c.salt
}

func (c *dummyConnInfo) FormatDSN() string {
	return "dummy@/dummy"
}

type dummyDBFuncs struct {
	prepareMetaTableErr   error
	existsMetaTableResult bool
	existsMetaTableErr    error
	readBindingResult     *databaseBindingRecord
	readBindingErr        error
	createBindingResult   *databaseBindingRecord
	createBindingErr      error
	deleteBindingErr      error
}

func (f *dummyDBFuncs) init() {
	f.prepareMetaTableErr = nil
	f.existsMetaTableResult = false
	f.existsMetaTableErr = nil
	f.readBindingResult = nil
	f.readBindingErr = nil
	f.createBindingResult = nil
	f.createBindingErr = nil
	f.deleteBindingErr = nil
}

func (f *dummyDBFuncs) databaseAPI() iaas.DatabaseAPI {
	return sacloudAPI.MariaDB()
}

func (f *dummyDBFuncs) buildConnInfo(host, dbName, user, password, salt string, port int) ConnectionInfo {
	return &dummyConnInfo{
		userName: user,
		password: password,
		host:     host,
		dbName:   dbName,
		salt:     salt,
		port:     port,
	}
}

func (f *dummyDBFuncs) prepareMetaTable(db *sql.DB, connInfo ConnectionInfo) error {
	return f.prepareMetaTableErr
}

func (f *dummyDBFuncs) existsMetaTable(db *sql.DB, connInfo ConnectionInfo) (bool, error) {
	return f.existsMetaTableResult, f.existsMetaTableErr
}

func (f *dummyDBFuncs) readBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	return f.readBindingResult, f.readBindingErr
}

func (f *dummyDBFuncs) createBinding(db *sql.DB, connInfo ConnectionInfo, bindingID string) (*databaseBindingRecord, error) {
	return f.createBindingResult, f.createBindingErr
}

func (f *dummyDBFuncs) deleteBinding(db *sql.DB, record *databaseBindingRecord) error {
	return f.deleteBindingErr
}

func TestGenericDBGetConn(t *testing.T) {

	s := &databaseHandler{
		serviceID: MariaDBServiceID,
		planID:    MariaDBPlan10GID,
		operation: operations.Provisioning,
	}
	conn, err := s.open(testConnInfo)
	defer conn.Close() // nolint

	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestDatabaseHandler_InstanceState(t *testing.T) {

	s := &databaseHandler{
		serviceID: MariaDBServiceID,
		planID:    MariaDBPlan10GID,
		operation: operations.Provisioning,
		dialect:   &dummyDBFuncs{},
	}

	t.Run("API returns 404 error", func(t *testing.T) {
		testDBAPI.readErr = apiError404
		defer func() {
			testDBAPI.readErr = nil
		}()

		state, err := s.InstanceState(instanceID)
		assert.Nil(t, state)
		assert.NoError(t, err)
	})

	t.Run("API return non 404 error", func(t *testing.T) {
		expectErr := errors.New("dummy")
		testDBAPI.readErr = expectErr
		defer func() {
			testDBAPI.readErr = nil
		}()

		state, err := s.InstanceState(instanceID)
		assert.Nil(t, state)
		assert.Error(t, err)
		assert.Equal(t, expectErr, err)
	})

	t.Run("API returns valid values", func(t *testing.T) {
		testDBAPI.readResult = mariaDB10GInstance(instanceID)
		defer func() {
			testDBAPI.readResult = nil
		}()

		state, err := s.InstanceState(instanceID)
		assert.NotNil(t, state)
		assert.NoError(t, err)
	})
}

func TestDatabaseHandler_BindingState(t *testing.T) {
	testDialect := &dummyDBFuncs{}
	s := &databaseHandler{
		serviceID: MariaDBServiceID,
		planID:    MariaDBPlan10GID,
		operation: operations.Binding,
		dialect:   testDialect,
	}
	testDBAPI.readResult = mariaDB10GInstance(instanceID)

	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("existsMetaTable func return nil with error if not created yet", func(t *testing.T) {
		testDialect.existsMetaTableErr = errors.New("dummy")
		defer testDialect.init()

		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		state, err := s.BindingState(instanceID, bindingID)
		assert.Nil(t, state)
		assert.Error(t, err)
	})

	t.Run("existsMetaTable func return nil with no error if not created yet", func(t *testing.T) {
		defer testDialect.init()
		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		state, err := s.BindingState(instanceID, bindingID)
		assert.Nil(t, state)
		assert.NoError(t, err)
	})

	t.Run("reading meta table returns no result", func(t *testing.T) {
		testDialect.existsMetaTableResult = true
		testDialect.readBindingErr = sql.ErrNoRows
		defer testDialect.init()

		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		state, err := s.BindingState(instanceID, bindingID)
		assert.Nil(t, state)
		assert.NoError(t, err)
	})

	t.Run("reading meta table returns error", func(t *testing.T) {
		testDialect.existsMetaTableResult = true
		testDialect.readBindingErr = errors.New("dummy")
		defer testDialect.init()

		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		state, err := s.BindingState(instanceID, bindingID)
		assert.Nil(t, state)
		assert.Error(t, err)
	})

	t.Run("reading meta table returns valid info", func(t *testing.T) {
		user, pass := "user", "pass"
		testDialect.existsMetaTableResult = true
		testDialect.readBindingResult = &databaseBindingRecord{
			bindingID: bindingID,
			username:  user,
			password:  pass,
		}
		defer testDialect.init()

		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		state, err := s.BindingState(instanceID, bindingID)
		assert.NoError(t, err)
		assert.NotNil(t, state)
		binding := state.Binding()
		assert.NotNil(t, binding)
		assert.NotNil(t, binding.Credentials)

		credential := binding.Credentials.(map[string]string)

		assert.Equal(t, credential["host"], connInfo.Host())
		assert.Equal(t, credential["port"], fmt.Sprintf("%d", connInfo.Port()))
		assert.Equal(t, credential["database"], user)
		assert.Equal(t, credential["password"], pass)
	})
}

func TestDatabaseHandler_CreateBinding(t *testing.T) {
	testDialect := &dummyDBFuncs{}
	s := &databaseHandler{
		serviceID: MariaDBServiceID,
		planID:    MariaDBPlan10GID,
		operation: operations.Binding,
		dialect:   testDialect,
	}
	testDBAPI.readResult = mariaDB10GInstance(instanceID)

	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("existsMetaTable func returns error", func(t *testing.T) {
		testDialect.existsMetaTableErr = errors.New("dummy")
		defer testDialect.init()

		binding, err := s.CreateBinding(instanceID, bindingID)
		assert.Nil(t, binding)
		assert.Error(t, err)
	})

	t.Run("readBinding func returns error", func(t *testing.T) {
		testDialect.existsMetaTableResult = true
		testDialect.readBindingErr = errors.New("dummy")
		defer testDialect.init()

		binding, err := s.CreateBinding(instanceID, bindingID)
		assert.Nil(t, binding)
		assert.Error(t, err)
	})

	t.Run("readBinding func returns BindingAlreadyExists error", func(t *testing.T) {
		testDialect.existsMetaTableResult = true
		testDialect.readBindingResult = &databaseBindingRecord{
			bindingID: bindingID,
			username:  "dummy",
			password:  "dummy",
		}
		defer testDialect.init()

		binding, err := s.CreateBinding(instanceID, bindingID)
		assert.Nil(t, binding)
		assert.Error(t, err)
		assert.IsType(t, &osb.BindingAlreadyExistsError{}, err)
	})

	t.Run("create valid binding", func(t *testing.T) {
		user, pass := "user", "pass"
		testDialect.existsMetaTableResult = true
		testDialect.createBindingResult = &databaseBindingRecord{
			bindingID: bindingID,
			username:  user,
			password:  pass,
		}
		defer testDialect.init()

		binding, err := s.CreateBinding(instanceID, bindingID)
		assert.NoError(t, err)
		assert.NotNil(t, binding)
		assert.NotNil(t, binding.Credentials)

		credential := binding.Credentials.(map[string]string)

		assert.Equal(t, credential["host"], connInfo.Host())
		assert.Equal(t, credential["port"], fmt.Sprintf("%d", connInfo.Port()))
		assert.Equal(t, credential["database"], user)
		assert.Equal(t, credential["password"], pass)
	})
}

func TestDatabaseHandler_DeleteBinding(t *testing.T) {
	testDialect := &dummyDBFuncs{}
	s := &databaseHandler{
		serviceID: MariaDBServiceID,
		planID:    MariaDBPlan10GID,
		operation: operations.Binding,
		dialect:   testDialect,
	}
	testDBAPI.readResult = mariaDB10GInstance(instanceID)

	t.Run("existsMetaTable func returns error", func(t *testing.T) {
		testDialect.existsMetaTableErr = errors.New("dummy")
		defer testDialect.init()

		err := s.DeleteBinding(instanceID, bindingID)
		assert.Error(t, err)
	})

	t.Run("readBinding func returns error", func(t *testing.T) {
		testDialect.existsMetaTableResult = true
		testDialect.readBindingErr = errors.New("dummy")
		defer testDialect.init()

		err := s.DeleteBinding(instanceID, bindingID)
		assert.Error(t, err)
	})

	t.Run("delete binding", func(t *testing.T) {
		user, pass := "user", "pass"
		record := &databaseBindingRecord{
			bindingID: bindingID,
			username:  user,
			password:  pass,
		}
		testDialect.existsMetaTableResult = true
		testDialect.readBindingResult = record
		defer testDialect.init()

		err := s.DeleteBinding(instanceID, bindingID)
		assert.NoError(t, err)
	})
}
