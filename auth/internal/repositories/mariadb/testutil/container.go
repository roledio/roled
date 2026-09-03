package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MariaDBTestContainer wraps the MariaDB testcontainer for testing
type MariaDBTestContainer struct {
	Container        *mariadb.MariaDBContainer
	ConnectionString string
}

// SetupMariaDB creates and starts a MariaDB container for testing.
// This should be called once per test suite (in TestMain or similar setup).
func SetupMariaDB(ctx context.Context, imageVersion string) (*MariaDBTestContainer, error) {
	// Default to 11.8 if no version specified
	if imageVersion == "" {
		imageVersion = "11.8"
	}

	// Create MariaDB container using the specialized module
	container, err := mariadb.Run(ctx,
		fmt.Sprintf("mariadb:%s", imageVersion),
		mariadb.WithDatabase("roled_test"),
		mariadb.WithUsername("root"),
		mariadb.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start MariaDB container: %w", err)
	}

	// Get connection string
	connectionString, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	return &MariaDBTestContainer{
		Container:        container,
		ConnectionString: connectionString,
	}, nil
}

// Terminate stops and removes the container
func (c *MariaDBTestContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}
