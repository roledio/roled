package databases

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/newrelic/go-agent/v3/integrations/nrmysql"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"github.com/roledio/roled/auth/pkg/utils/validationutil"
)

func Open(config Config) (*sql.DB, error) {
	err := validationutil.ValidateStruct(context.Background(), config)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(config.GetDriver(), config.GetDSN())
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
	return db, nil
}
