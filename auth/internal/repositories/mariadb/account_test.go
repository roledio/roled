package mariadb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccountRepository_FindByID tests the FindByID method
func TestAccountRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("returns account when found", func(t *testing.T) {
		// Clean table before test
		testSuite.CleanTables(t, "accounts")

		// Create test account
		fixture := testutil.AccountFixture{
			ID:          "acc_find_by_id_test",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		expected, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err, "Failed to create test account")

		// Test FindByID
		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err, "FindByID should not return error")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions
		assert.Equal(t, expected.ID, result.ID, "ID should match")
		assert.Equal(t, expected.Name, result.Name, "Name should match")
		assert.Equal(t, expected.Description, result.Description, "Description should match")
		assert.Equal(t, expected.IsActive, result.IsActive, "IsActive should match")
		assert.Equal(t, expected.IsSystem, result.IsSystem, "IsSystem should match")
		assert.Nil(t, result.DeletedAt, "DeletedAt should be nil")
	})

	t.Run("returns nil when account not found", func(t *testing.T) {
		// Clean table before test
		testSuite.CleanTables(t, "accounts")

		// Test FindByID with non-existent ID
		result, err := repo.FindByID(ctx, "non_existent_id")
		require.NoError(t, err, "FindByID should not return error for non-existent ID")
		assert.Nil(t, result, "Result should be nil for non-existent ID")
	})

	t.Run("excludes soft-deleted accounts", func(t *testing.T) {
		// Clean table before test
		testSuite.CleanTables(t, "accounts")

		// Create test account
		fixture := testutil.AccountFixture{
			ID:          "acc_soft_deleted",
			Name:        "Soft Deleted Account",
			Description: "This account will be soft deleted",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err, "Failed to create test account")

		// Soft delete the account
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE accounts SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err, "Failed to soft delete account")

		// Test FindByID - should not find soft-deleted account
		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err, "FindByID should not return error")
		assert.Nil(t, result, "Result should be nil for soft-deleted account")
	})

	t.Run("handles different account types", func(t *testing.T) {
		// Clean table before test
		testSuite.CleanTables(t, "accounts")

		tests := []struct {
			name    string
			fixture testutil.AccountFixture
		}{
			{
				name: "system account",
				fixture: testutil.AccountFixture{
					ID:          "acc_system_type",
					Name:        "System Account",
					Description: "System managed",
					IsActive:    true,
					IsSystem:    true,
				},
			},
			{
				name: "inactive account",
				fixture: testutil.AccountFixture{
					ID:          "acc_inactive_type",
					Name:        "Inactive Account",
					Description: "Inactive account",
					IsActive:    false,
					IsSystem:    false,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				expected, err := testutil.CreateAccount(ctx, testSuite.GetDB(), tt.fixture)
				require.NoError(t, err, "Failed to create test account")

				result, err := repo.FindByID(ctx, tt.fixture.ID)
				require.NoError(t, err)
				require.NotNil(t, result)

				assert.Equal(t, expected.ID, result.ID)
				assert.Equal(t, expected.IsActive, result.IsActive)
				assert.Equal(t, expected.IsSystem, result.IsSystem)
			})
		}
	})
}

// TestAccountRepository_Create tests the Create method
func TestAccountRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("creates account successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		account := &entities.Account{
			ID:          "acc_create_test",
			Name:        "New Account",
			Description: "New account description",
			IsActive:    true,
			IsSystem:    false,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		err := repo.Create(ctx, account)
		require.NoError(t, err, "Create should not return error")

		// Verify the account was created
		result, err := repo.FindByID(ctx, account.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, account.ID, result.ID)
		assert.Equal(t, account.Name, result.Name)
		assert.Equal(t, account.Description, result.Description)
		assert.Equal(t, account.IsActive, result.IsActive)
		assert.Equal(t, account.IsSystem, result.IsSystem)
	})

	t.Run("creates system account", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		account := &entities.Account{
			ID:          "acc_create_system",
			Name:        "System Account",
			Description: "System account description",
			IsActive:    true,
			IsSystem:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		err := repo.Create(ctx, account)
		require.NoError(t, err, "Create should not return error")

		// Verify the account was created as system account
		result, err := repo.FindByID(ctx, account.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsSystem)
	})

	t.Run("fails with duplicate ID", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		account := &entities.Account{
			ID:          "acc_duplicate_test",
			Name:        "Duplicate Account",
			Description: "This will be duplicated",
			IsActive:    true,
			IsSystem:    false,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		// Create first time - should succeed
		err := repo.Create(ctx, account)
		require.NoError(t, err)

		// Try to create with same ID - should fail
		err = repo.Create(ctx, account)
		assert.Error(t, err, "Create with duplicate ID should return error")
	})
}

// TestAccountRepository_Update tests the Update method
func TestAccountRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("updates account successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create test account
		fixture := testutil.AccountFixture{
			ID:          "acc_update_test",
			Name:        "Original Name",
			Description: "Original Description",
			IsActive:    false,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Update the account
		updatedAccount := &entities.Account{
			ID:          fixture.ID,
			Name:        "Updated Name",
			Description: "Updated Description",
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedAccount)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify the update
		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Name", result.Name)
		assert.Equal(t, "Updated Description", result.Description)
		assert.True(t, result.IsActive)
	})

	t.Run("returns zero rows when account not found", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		updatedAccount := &entities.Account{
			ID:          "non_existent_id",
			Name:        "Updated Name",
			Description: "Updated Description",
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedAccount)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should return 0 rows affected for non-existent account")
	})

	t.Run("does not update soft-deleted accounts", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create and soft-delete an account
		fixture := testutil.AccountFixture{
			ID:          "acc_update_deleted",
			Name:        "Deleted Account",
			Description: "This will be deleted",
			IsActive:    false,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE accounts SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		// Try to update the soft-deleted account
		updatedAccount := &entities.Account{
			ID:          fixture.ID,
			Name:        "Updated Name",
			Description: "Updated Description",
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedAccount)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should not update soft-deleted account")
	})
}

// TestAccountRepository_DeleteByID tests the DeleteByID method
func TestAccountRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("soft deletes account successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create test account
		fixture := testutil.AccountFixture{
			ID:          "acc_delete_test",
			Name:        "Account to Delete",
			Description: "This will be deleted",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Delete the account
		rowsAffected, err := repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify it's soft deleted (FindByID should return nil)
		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Nil(t, result, "Deleted account should not be found by FindByID")

		// Verify it still exists in database with deleted_at set
		var deletedAt *time.Time
		err = testSuite.GetDB().GetContext(ctx, &deletedAt, "SELECT deleted_at FROM accounts WHERE id = ?", fixture.ID)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt, "deleted_at should be set")
	})

	t.Run("returns zero rows when account not found", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		rowsAffected, err := repo.DeleteByID(ctx, "non_existent_id")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should return 0 rows affected for non-existent account")
	})

	t.Run("does not delete already deleted account", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create and soft-delete an account
		fixture := testutil.AccountFixture{
			ID:          "acc_double_delete",
			Name:        "Double Delete",
			Description: "This will be deleted twice",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// First delete
		rowsAffected, err := repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Second delete - should return 0 rows
		rowsAffected, err = repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should not delete already deleted account")
	})
}

// TestAccountRepository_FindAll tests the FindAll method
func TestAccountRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("returns all accounts with default pagination", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create test accounts
		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Test FindAll with default pagination
		req := &models.GetAccountsRequest{}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)
		assert.Len(t, result, 3, "Should return all 3 accounts")
	})

	t.Run("filters by is_active", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		isActive := true
		req := &models.GetAccountsRequest{IsActive: &isActive}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1, "Should return at least 1 active account")

		for _, acc := range result {
			assert.True(t, acc.IsActive, "All returned accounts should be active")
		}
	})

	t.Run("searches by name", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		req := &models.GetAccountsRequest{Search: "System"}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1, "Should return at least 1 account with 'System' in name")

		for _, acc := range result {
			assert.Contains(t, acc.Name, "System", "Returned account name should contain 'System'")
		}
	})

	t.Run("filters by filterAccountID", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		accounts, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Find a non-system account
		var targetID string
		for _, acc := range accounts {
			if !acc.IsSystem {
				targetID = acc.ID
				break
			}
		}

		req := &models.GetAccountsRequest{}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, &targetID)
		require.NoError(t, err)
		assert.Len(t, result, 1, "Should return exactly 1 account")
		assert.Equal(t, targetID, result[0].ID, "Should return the filtered account")
		assert.False(t, result[0].IsSystem, "Should not return system accounts when filtering")
	})

	t.Run("supports pagination", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create more accounts for pagination test
		for i := 0; i < 15; i++ {
			fixture := testutil.AccountFixture{
				ID:          fmt.Sprintf("acc_page_%02d", i),
				Name:        fmt.Sprintf("Pagination Account %d", i),
				Description: "Pagination test",
				IsActive:    true,
				IsSystem:    false,
			}
			_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		// Test first page
		req := &models.GetAccountsRequest{
			PageRequest: pkgmodels.PageRequest{
				PageNum:  1,
				PageSize: 5,
			},
		}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)
		assert.Len(t, result, 5, "Should return 5 accounts for page 1")

		// Test second page
		req.PageNum = 2
		result2, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)
		assert.Len(t, result2, 5, "Should return 5 accounts for page 2")

		// Verify different accounts on different pages
		assert.NotEqual(t, result[0].ID, result2[0].ID, "Pages should have different accounts")
	})

	t.Run("excludes soft-deleted accounts", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create account and soft-delete it
		fixture := testutil.AccountFixture{
			ID:          "acc_findall_deleted",
			Name:        "Deleted Account",
			Description: "Should not appear",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Soft delete
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE accounts SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		req := &models.GetAccountsRequest{}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req, nil)
		require.NoError(t, err)

		for _, acc := range result {
			assert.NotEqual(t, fixture.ID, acc.ID, "Deleted account should not be in results")
		}
	})
}

// TestAccountRepository_Count tests the Count method
func TestAccountRepository_Count(t *testing.T) {
	ctx := context.Background()
	repo := NewAccountRepository(testSuite.GetDB())

	t.Run("counts all accounts", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		req := &models.GetAccountsRequest{}
		count, err := repo.Count(ctx, req, nil)
		require.NoError(t, err)
		assert.Equal(t, 3, count, "Should count all 3 accounts")
	})

	t.Run("counts with is_active filter", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		isActive := true
		req := &models.GetAccountsRequest{IsActive: &isActive}

		count, err := repo.Count(ctx, req, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Should count at least 1 active account")
	})

	t.Run("counts with search filter", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		_, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		req := &models.GetAccountsRequest{Search: "System"}

		count, err := repo.Count(ctx, req, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1, "Should count at least 1 account with 'System'")
	})

	t.Run("counts with filterAccountID", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		fixtures := testutil.DefaultAccountFixtures()
		accounts, err := testutil.CreateAccounts(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Find a non-system account
		var targetID string
		for _, acc := range accounts {
			if !acc.IsSystem {
				targetID = acc.ID
				break
			}
		}

		req := &models.GetAccountsRequest{}
		count, err := repo.Count(ctx, req, &targetID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should count exactly 1 account")
	})

	t.Run("excludes soft-deleted accounts from count", func(t *testing.T) {
		testSuite.CleanTables(t, "accounts")

		// Create account and soft-delete it
		fixture := testutil.AccountFixture{
			ID:          "acc_count_deleted",
			Name:        "Deleted Account",
			Description: "Should not be counted",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Soft delete
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE accounts SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		req := &models.GetAccountsRequest{}
		count, err := repo.Count(ctx, req, nil)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Should not count deleted accounts")
	})
}
