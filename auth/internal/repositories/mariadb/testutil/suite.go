package testutil

import (
	"context"
	"testing"
)

// TestSuite manages the test container and database for all repository tests.
// It creates the container once and reuses it across all tests.
type TestSuite struct {
	Container *MariaDBTestContainer
	DB        *TestDB
	Ctx       context.Context
}

// CleanTables cleans specified tables between tests
func (s *TestSuite) CleanTables(t *testing.T, tables ...string) {
	t.Helper()
	if err := s.DB.CleanTables(s.Ctx, tables...); err != nil {
		t.Fatalf("Failed to clean tables: %v", err)
	}
}

// GetDB returns the database connection for use in tests
func (s *TestSuite) GetDB() *TestDB {
	return s.DB
}

// GetContainer returns the MariaDB container for inspection
func (s *TestSuite) GetContainer() *MariaDBTestContainer {
	return s.Container
}
