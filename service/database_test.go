package service

import (
	"database/sql"
	"database/sql/driver"
)

type dummyDriver struct{}

func (d *dummyDriver) Open(name string) (driver.Conn, error) {
	return &dummyConn{}, nil
}

type dummyConn struct{}

func (c *dummyConn) Prepare(query string) (driver.Stmt, error) {
	return nil, nil
}
func (c *dummyConn) Close() error {
	return nil
}
func (c *dummyConn) Begin() (driver.Tx, error) {
	return nil, nil
}

func initTestDatabase(driver, dsn string, execSQLs ...string) error { // nolint

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close() // nolint

	for _, query := range execSQLs {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}
