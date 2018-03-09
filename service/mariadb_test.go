package service

import (
	"fmt"
	"testing"

	"context"
	"database/sql"
	"errors"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/broker/operations"
	"github.com/stretchr/testify/assert"
	"time"
)

const (
	testMariaDBRootPassword = "passwo@d"
	testMariaDBUser         = "testuser"
	testMariaDBPassword     = "passwo@d"
)

func getMariaDBHandler(operation string, paramJSON string) *databaseHandler {
	return newMariaDBServiceHandler(
		operation,
		MariaDBServiceID,
		MariaDBPlan10GID,
		[]byte(paramJSON),
	)
}

var (
	instanceID             = "xxxx"
	bindingID              = "xxxx"
	mariaDBTestSwitchID    = 999999999999
	validProvisioningParam = `
    	{
    		"switchID" : 123456789012,
    		"ipaddress" : "192.2.0.10",
    		"maskLen" : 24,
    		"defaultRoute" : "192.2.0.1"
	    }`
)

func mariaDB10GInstance(instanceID string) *sacloud.Database {
	p := sacloud.NewCreateMariaDBDatabaseValue()
	p.Plan = sacloud.DatabasePlan10G
	p.SwitchID = fmt.Sprintf("%d", mariaDBTestSwitchID)
	p.IPAddress1 = "192.2.0.10"
	p.MaskLen = 24
	p.DefaultRoute = "192.2.0.1"
	p.Name = instanceID
	return sacloud.CreateNewDatabase(p)
}

func mariaDBTestConnInstance(instanceID string) *sacloud.Database {
	p := sacloud.NewCreateMariaDBDatabaseValue()
	p.Plan = sacloud.DatabasePlan10G
	p.SwitchID = fmt.Sprintf("%d", mariaDBTestSwitchID)
	p.DefaultUser = testMariaDBUser
	p.UserPassword = testMariaDBPassword
	p.IPAddress1 = "127.0.0.1"
	p.MaskLen = 24
	p.DefaultRoute = "127.0.0.1"
	p.Name = instanceID
	return sacloud.CreateNewDatabase(p)
}

func TestMariaDBServiceValidate(t *testing.T) {
	t.Run("Provisioning", func(t *testing.T) {
		t.Run("Empty rawParameter", func(t *testing.T) {
			s := getMariaDBHandler(operations.Provisioning, ``)
			result, err := s.IsValid()
			assert.False(t, result)
			assert.Error(t, err)
			assert.Nil(t, s.parameter)
		})

		t.Run("Valid rawParameter", func(t *testing.T) {
			s := getMariaDBHandler(operations.Provisioning, validProvisioningParam)
			result, err := s.IsValid()
			assert.True(t, result)
			assert.NoError(t, err)
			assert.NotNil(t, s.parameter)
		})
	})
	t.Run("Binding", func(t *testing.T) {
		t.Run("Nil rawParameter", func(t *testing.T) {
			s := getMariaDBHandler(operations.Binding, ``)
			result, err := s.IsValid()
			assert.True(t, result)
			assert.NoError(t, err)
			assert.Nil(t, s.parameter)
		})
	})

	t.Run("Other operations", func(t *testing.T) {
		s := getMariaDBHandler(operations.Unbinding, ``)
		result, err := s.IsValid()
		assert.False(t, result)
		assert.Error(t, err)
		assert.Nil(t, s.parameter)
	})
}

func TestMariaDBHandler_prepareMetaTable(t *testing.T) {
	if !existsTestEnvVars("TEST_DB") {
		t.Skipf("environment variable %q is empty. skip.", "TEST_DB")
		return
	}

	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: mariaDBTestConnInstance(instanceID), // connect to localhost
		},
	}

	cleanup := initMariaDB()
	defer cleanup()

	s := getMariaDBHandler(operations.Binding, ``)
	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create metadata table if not exists", func(t *testing.T) {
		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		// pre: meta table is not exists
		exists, err := s.dialect.existsMetaTable(db, connInfo)
		assert.False(t, exists)
		assert.NoError(t, err)

		err = s.dialect.prepareMetaTable(db, connInfo)
		assert.NoError(t, err)

		// post
		exists, err = s.dialect.existsMetaTable(db, connInfo)
		assert.True(t, exists)
		assert.NoError(t, err)
	})
}

func TestMariaDBHandler_create_read_delete(t *testing.T) {
	if !existsTestEnvVars("TEST_DB") {
		t.Skipf("environment variable %q is empty. skip.", "TEST_DB")
		return
	}

	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: mariaDBTestConnInstance(instanceID), // connect to localhost
		},
	}

	cleanup := initMariaDB()
	defer cleanup()

	s := getMariaDBHandler(operations.Binding, ``)
	// collect db-info
	connInfo, err := s.connInfo(instanceID)
	if err != nil {
		t.Fatal(err)
	}

	var createdRecord *databaseBindingRecord
	t.Run("create user database and binding record", func(t *testing.T) {
		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		err = s.dialect.prepareMetaTable(db, connInfo)
		assert.NoError(t, err)

		record, err := s.dialect.createBinding(db, connInfo, bindingID)
		assert.NotNil(t, record)
		assert.NoError(t, err)

		// check metatable exists
		exists, err := s.dialect.existsMetaTable(db, connInfo)
		assert.True(t, exists)
		assert.NoError(t, err)

		createdRecord = record
	})

	t.Run("read binding", func(t *testing.T) {
		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		record, err := s.dialect.readBinding(db, connInfo, bindingID)
		assert.NoError(t, err)
		assert.NotNil(t, record)

		assert.EqualValues(t, createdRecord, record)
	})

	t.Run("delete binding", func(t *testing.T) {
		// connect db
		db, err := s.open(connInfo)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		err = s.dialect.deleteBinding(db, createdRecord)
		assert.NoError(t, err)

		// check delete
		record, err := s.dialect.readBinding(db, connInfo, bindingID)
		assert.NoError(t, err)
		assert.Nil(t, record)

	})
}

func TestMariaDBConnInfo(t *testing.T) {
	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: mariaDB10GInstance(instanceID),
		},
	}
	s := getMariaDBHandler(operations.Provisioning, validProvisioningParam)

	info, err := s.connInfo(instanceID)

	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func initMariaDB(queries ...string) func() {

	cleanup, err := startDocker("mariadb:10.2",
		map[string]string{
			"MYSQL_ROOT_PASSWORD": testMariaDBRootPassword,
			"MYSQL_DATABASE":      testMariaDBUser,
			"MYSQL_USER":          testMariaDBUser,
			"MYSQL_PASSWORD":      testMariaDBPassword,
		},
		[]int{3306},
		"",
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res := make(chan bool)
	go func() {
		for {
			db, err := sql.Open("mysql",
				fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s",
					"root", testMariaDBRootPassword, testMariaDBUser),
			)
			if err == nil {
				_, err = db.Query("SELECT 1 FROM INFORMATION_SCHEMA.TABLES LIMIT 1")
				if err == nil {
					_, err = db.Exec(
						fmt.Sprintf("GRANT ALL ON *.* TO '%s'@'%%' WITH GRANT OPTION", testMariaDBUser),
					)
					if err == nil {
						db.Close()
						res <- true
						return
					}
					panic(err)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-res:
	case <-ctx.Done():
		panic(errors.New("sql.Open is timed out"))
	}

	return cleanup
}
