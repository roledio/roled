package databases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/pressly/goose/v3"
)

const (
	DialectMySQL    = "mysql"
	DialectPostgres = "postgres"
)

func Migrate(db *sql.DB, dialect string, dir string) error {
	if db == nil {
		log.Error("Database connection is nil, cannot run migration")
		return fmt.Errorf("database connection is nil")
	}
	if err := goose.SetDialect(dialect); err != nil {
		log.Error("Set dialect error: ", err)
		return err
	}
	if err := goose.RunContext(context.Background(), "up", db, dir); err != nil {
		if errors.Is(err, goose.ErrNoMigrationFiles) || strings.HasSuffix(err.Error(), "directory does not exist") {
			log.Info("No migration directory or files found, skipping migration (nothing to do)")
			return nil
		}
		log.Error("Migration up error: ", err)
		return err
	}
	return nil
}
