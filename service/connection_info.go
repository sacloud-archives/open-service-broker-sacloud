package service

// ConnectionInfo represents info to connect database
type ConnectionInfo interface {
	DriverName() string
	UserName() string
	Password() string
	Host() string
	Port() int
	DBName() string
	Salt() string
	FormatDSN() string
}
