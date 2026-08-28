package databases

import "fmt"

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

type Config struct {
	// Connection URL
	DSN string

	// Database driver name
	Driver string `validate:"required,oneof=mysql postgres"`

	// Database host
	Host string `validate:"required_if=DSN ''"`

	// Database username
	Username string `validate:"required_if=DSN ''"`

	// Database password
	Password string `validate:"required_if=DSN ''"`

	// Database name
	Name string `validate:"required_if=DSN ''"`

	// Database port
	Port int64 `validate:"required_if=DSN ''"`

	// Postgres SSL mode
	SSLMode string `validate:"omitempty,oneof=disable require verify-ca verify-full"`

	MaxOpenConns int

	MaxIdleConns int

	// Enable NewRelic integration
	Newrelic bool

	// Application name
	ApplicationName string
}

// GetDSN returns the Data Source Name for the database connection specified in the Config.
// If DSN is already provided, it returns that directly. Otherwise, it constructs the DSN based on
// the Driver and other connection parameters.
func (c *Config) GetDSN() string {
	dsn := c.DSN
	if dsn != "" {
		return dsn
	}
	switch c.Driver {
	case DriverMySQL:
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Username,
			c.Password,
			c.Host,
			c.Port,
			c.Name,
		)
	case DriverPostgres:
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
			c.Host,
			c.Port,
			c.Username,
			c.Password,
			c.Name,
		)
		if c.ApplicationName != "" {
			dsn += " application_name='" + c.ApplicationName + "'"
		}
		if c.SSLMode != "" {
			dsn += " sslmode=" + c.SSLMode
		}
	}
	return dsn
}

// GetDriver returns the database driver name, taking into account newrelic integration.
func (c *Config) GetDriver() string {
	driver := c.Driver
	if c.Driver == DriverMySQL && c.Newrelic {
		driver = "nrmysql"
	} else if c.Driver == DriverPostgres && c.Newrelic {
		driver = "nrpostgres"
	}
	return driver
}
