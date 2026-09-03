package mariadb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	"github.com/roledio/roled/auth/pkg/utils/idutil"

	// Import migrations to register Go migrations
	_ "github.com/roledio/roled/auth/migrations"
)

// testSuite is the shared test suite for all repository tests
var testSuite *testutil.TestSuite

// migrationsPath is the fixed path to the migrations directory
const migrationsPath = "../../../migrations"

// TestMain sets up the test suite once for all tests in this package
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup environment variables requied for migrations
	_ = os.Setenv("BASE_URL", "http://localhost:8080")
	_ = os.Setenv("CONSOLE_BASE_URL", "http://localhost:4000")
	_ = os.Setenv("ENCRYPTION_MASTER_KEY", idutil.NanoID(32))

	fmt.Println("Setting up MariaDB test container...")

	// Setup MariaDB container with version 11.8
	container, err := testutil.SetupMariaDB(ctx, "11.8")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup MariaDB container: %v\n", err)
		os.Exit(1)
	}

	// Create database connection
	db, err := testutil.NewTestDB(container.ConnectionString)
	if err != nil {
		_ = container.Terminate(ctx)
		fmt.Fprintf(os.Stderr, "Failed to create test database: %v\n", err)
		os.Exit(1)
	}

	// Run migrations using databases.Migrate (includes both SQL and Go migrations)
	if err := db.RunMigrations(migrationsPath); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Test suite setup complete")

	// Create test suite
	testSuite = &testutil.TestSuite{
		Container: container,
		DB:        db,
		Ctx:       ctx,
	}

	// Run tests
	code := m.Run()

	// Cleanup
	fmt.Println("Cleaning up test suite...")
	_ = db.Close()
	_ = container.Terminate(ctx)
	fmt.Println("Test suite cleanup complete")

	os.Exit(code)
}
