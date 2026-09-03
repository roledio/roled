package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
)

// AccountFixture represents test data for account
type AccountFixture struct {
	ID          string
	Name        string
	Description string
	IsActive    bool
	IsSystem    bool
}

// CreateAccount inserts an account into the database and returns the entity
func CreateAccount(ctx context.Context, db *TestDB, fixture AccountFixture) (*entities.Account, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	account := &entities.Account{
		ID:          fixture.ID,
		Name:        fixture.Name,
		Description: fixture.Description,
		IsActive:    fixture.IsActive,
		IsSystem:    fixture.IsSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `INSERT INTO accounts (id, name, description, is_active, is_system, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		account.ID,
		account.Name,
		account.Description,
		account.IsActive,
		account.IsSystem,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// CreateAccounts inserts multiple accounts into the database
func CreateAccounts(ctx context.Context, db *TestDB, fixtures []AccountFixture) ([]entities.Account, error) {
	accounts := make([]entities.Account, 0, len(fixtures))
	for _, f := range fixtures {
		account, err := CreateAccount(ctx, db, f)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}

// DefaultAccountFixtures returns common test fixtures for accounts
func DefaultAccountFixtures() []AccountFixture {
	return []AccountFixture{
		{
			ID:          "acc_001_test_account_1",
			Name:        "Test Account 1",
			Description: "First test account",
			IsActive:    true,
			IsSystem:    false,
		},
		{
			ID:          "acc_002_test_account_2",
			Name:        "Test Account 2",
			Description: "Second test account",
			IsActive:    false,
			IsSystem:    false,
		},
		{
			ID:          "acc_003_system_account",
			Name:        "System Account",
			Description: "System managed account",
			IsActive:    true,
			IsSystem:    true,
		},
	}
}
