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
	testPostgreSQLUser     = "testuser"
	testPostgreSQLPassword = "passwo@d"
)

func getPostgreSQLHandler(operation string, paramJSON string) *databaseHandler {
	return newPostgreSQLServiceHandler(
		operation,
		PostgreSQLServiceID,
		MariaDBPlan10GID,
		[]byte(paramJSON),
	)
}

var postgreSQLTestSwitchID = 999999999999

func postgreSQLTestConnInstance(instanceID string) *sacloud.Database {
	p := sacloud.NewCreatePostgreSQLDatabaseValue()
	p.Plan = sacloud.DatabasePlan10G
	p.SwitchID = fmt.Sprintf("%d", postgreSQLTestSwitchID)
	p.DefaultUser = testPostgreSQLUser
	p.UserPassword = testPostgreSQLPassword
	p.IPAddress1 = "127.0.0.1"
	p.MaskLen = 24
	p.DefaultRoute = "127.0.0.1"
	p.Name = instanceID
	return sacloud.CreateNewDatabase(p)
}

func TestPostgreSQLServiceValidate(t *testing.T) {
	t.Run("Provisioning", func(t *testing.T) {
		t.Run("Empty rawParameter", func(t *testing.T) {
			s := getPostgreSQLHandler(operations.Provisioning, ``)
			result, err := s.IsValid()
			assert.False(t, result)
			assert.Error(t, err)
			assert.Nil(t, s.parameter)
		})

		t.Run("Valid rawParameter", func(t *testing.T) {
			s := getPostgreSQLHandler(operations.Provisioning, validProvisioningParam)
			result, err := s.IsValid()
			assert.True(t, result)
			assert.NoError(t, err)
			assert.NotNil(t, s.parameter)
		})
	})
	t.Run("Binding", func(t *testing.T) {
		t.Run("Nil rawParameter", func(t *testing.T) {
			s := getPostgreSQLHandler(operations.Binding, ``)
			result, err := s.IsValid()
			assert.True(t, result)
			assert.NoError(t, err)
			assert.Nil(t, s.parameter)
		})
	})

	t.Run("Other operations", func(t *testing.T) {
		s := getPostgreSQLHandler(operations.Unbinding, ``)
		result, err := s.IsValid()
		assert.False(t, result)
		assert.Error(t, err)
		assert.Nil(t, s.parameter)
	})
}

func TestPostgreSQLHandler_prepareMetaTable(t *testing.T) {
	if !existsTestEnvVars("TEST_DB") {
		t.Skipf("Environemnt variable %q is empty. skip.", "TEST_DB")
		return
	}

	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: postgreSQLTestConnInstance(instanceID), // connect to localhost
		},
	}

	cleanup := initPostgreSQL()
	defer cleanup()

	s := getPostgreSQLHandler(operations.Binding, ``)
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

func TestPostgreSQLHandler_create_read_delete(t *testing.T) {
	if !existsTestEnvVars("TEST_DB") {
		t.Skipf("Environemnt variable %q is empty. skip.", "TEST_DB")
		return
	}

	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: postgreSQLTestConnInstance(instanceID), // connect to localhost
		},
	}

	cleanup := initPostgreSQL()
	defer cleanup()

	s := getPostgreSQLHandler(operations.Binding, ``)
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

func TestPostgreSQLConnInfo(t *testing.T) {
	sacloudAPI = &dummyAPI{
		dbAPI: &genericDBDummyAPI{
			readResult: postgreSQLTestConnInstance(instanceID),
		},
	}
	s := getPostgreSQLHandler(operations.Provisioning, validProvisioningParam)

	info, err := s.connInfo(instanceID)

	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func initPostgreSQL(queries ...string) func() {

	cleanup, err := startDocker("postgres:9.6",
		map[string]string{
			"POSTGRES_DB":       testPostgreSQLUser,
			"POSTGRES_USER":     testPostgreSQLUser,
			"POSTGRES_PASSWORD": testPostgreSQLPassword,
		},
		[]int{5432},
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
			db, err := sql.Open("postgres",
				fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable",
					testPostgreSQLUser, testPostgreSQLPassword, testPostgreSQLUser),
			)
			if err == nil {
				_, err = db.Query("SELECT 1 FROM INFORMATION_SCHEMA.TABLES LIMIT 1")
				if err == nil {
					db.Close()
					res <- true
					return
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
