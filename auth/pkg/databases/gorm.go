package databases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/newrelic/go-agent/v3/integrations/nrmysql"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"github.com/roledio/roled/auth/pkg/utils/validationutil"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (c *Config) dsn() string {
	dsn := c.DSN
	if dsn != "" {
		return dsn
	}
	if c.Driver == DriverMySQL {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Username,
			c.Password,
			c.Host,
			c.Port,
			c.Name,
		)
	} else if c.Driver == DriverPostgres {
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

func (c *Config) driver() string {
	driver := c.Driver
	if c.Driver == DriverMySQL && c.Newrelic {
		driver = "nrmysql"
	} else if c.Driver == DriverPostgres && c.Newrelic {
		driver = "nrpostgres"
	}
	return driver
}

func (c *Config) dialector(db *sql.DB) gorm.Dialector {
	var d gorm.Dialector
	switch c.Driver {
	case DriverMySQL:
		d = mysql.New(mysql.Config{Conn: db})
	case DriverPostgres:
		d = postgres.New(postgres.Config{Conn: db})
	}
	return d
}

// Connect creates a new GORM database connection to database using the provided configuration.
func Connect(config Config) (*gorm.DB, error) {
	err := validationutil.ValidateStruct(context.Background(), config)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(config.driver(), config.dsn())
	if err != nil {
		return nil, err
	}

	// Setting the connection pool.
	// https://medium.com/propertyfinder-engineering/go-and-mysql-setting-up-connection-pooling-4b778ef8e560#0ffb
	// https://go.dev/doc/database/manage-connections
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verify that the data source name is valid. sql.Open() may just validate its arguments without creating a connection to the database.
	// A context with a 5 second timeout is made to ensure that the program doesn’t get stuck when pinging the DB
	// in case there is a network error or any other error.
	// https://pkg.go.dev/database/sql#Open
	// https://golangbot.com/connect-create-db-mysql/#pinging-the-db
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	gc := &gorm.Config{}
	if config.Logger != nil {
		gc.Logger = config.Logger
	}
	return gorm.Open(config.dialector(db), gc)
}
