package testutil

import (
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/roledio/roled/auth/pkg/databases"
)

// TestDB holds the database connection for testing
type TestDB struct {
	*sqlx.DB
}

// NewTestDB creates a new database connection for testing
func NewTestDB(connectionString string) (*TestDB, error) {
	db, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings for tests
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	return &TestDB{db}, nil
}

// RunMigrations runs database migrations using the databases.Migrate function.
// The migrationsPath should be the relative path from where the test is run.
func (db *TestDB) RunMigrations(migrationsPath string) error {
	// Use the same Migrate function as the main application
	if err := databases.Migrate(db.DB.DB, databases.DialectMySQL, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// CleanTables truncates all tables in the database (for cleanup between tests)
func (db *TestDB) CleanTables(ctx context.Context, tables ...string) error {
	// Disable foreign key checks temporarily
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", table)); err != nil {
			// Re-enable foreign key checks even on error
			_, _ = db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *TestDB) Close() error {
	return db.DB.Close()
}
