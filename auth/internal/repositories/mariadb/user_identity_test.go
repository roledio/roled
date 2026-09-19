package mariadb

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserIdentityRepository_Create tests the Create method
func TestUserIdentityRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewUserIdentityRepository(testSuite.GetDB())

	t.Run("creates user identity successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account first
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_123",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		userIdentity := &entities.UserIdentity{
			ID:             "uid_create_test",
			UserID:         "user_123",
			Provider:       "github",
			ProviderUserID: "github_user_456",
			ProjectID:      "proj_123",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		err = repo.Create(ctx, userIdentity)
		require.NoError(t, err, "Create should not return error")

		// Verify the user identity was created
		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_456", "proj_123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, userIdentity.ID, result.ID)
		assert.Equal(t, userIdentity.UserID, result.UserID)
		assert.Equal(t, userIdentity.Provider, result.Provider)
		assert.Equal(t, userIdentity.ProviderUserID, result.ProviderUserID)
		assert.Equal(t, userIdentity.ProjectID, result.ProjectID)
	})

	t.Run("creates user identity with different provider", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_456",
			Name:        "Test Account 2",
			Description: "Test Description 2",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_456",
			AccountID: "acc_456",
			Name:      "Test Project 2",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test2@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_789",
			AccountID: "acc_456",
			ProjectID: "proj_456",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		userIdentity := &entities.UserIdentity{
			ID:             "uid_create_google",
			UserID:         "user_789",
			Provider:       "google",
			ProviderUserID: "google_user_abc",
			ProjectID:      "proj_456",
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		err = repo.Create(ctx, userIdentity)
		require.NoError(t, err, "Create should not return error")

		// Verify the user identity was created
		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "google", "google_user_abc", "proj_456")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "google", result.Provider)
	})
}

// TestUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID tests the FindByProviderAndProviderUserIDAndProjectID method
func TestUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserIdentityRepository(testSuite.GetDB())

	t.Run("returns user identity when found", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_123",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixture := testutil.UserIdentityFixture{
			ID:             "uid_find_test",
			UserID:         "user_123",
			Provider:       "github",
			ProviderUserID: "github_user_789",
			ProjectID:      "proj_123",
		}
		expected, err := testutil.CreateUserIdentity(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err, "Failed to create test user identity")

		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_789", "proj_123")
		require.NoError(t, err, "FindByProviderAndProviderUserIDAndProjectID should not return error")
		require.NotNil(t, result, "Result should not be nil")

		assert.Equal(t, expected.ID, result.ID)
		assert.Equal(t, expected.UserID, result.UserID)
		assert.Equal(t, expected.Provider, result.Provider)
		assert.Equal(t, expected.ProviderUserID, result.ProviderUserID)
		assert.Equal(t, expected.ProjectID, result.ProjectID)
	})

	t.Run("returns nil when user identity not found", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users")

		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "non_existent_user", "proj_123")
		require.NoError(t, err, "FindByProviderAndProviderUserIDAndProjectID should not return error for non-existent user identity")
		assert.Nil(t, result, "Result should be nil for non-existent user identity")
	})

	t.Run("excludes soft-deleted user identities", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_456",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixture := testutil.UserIdentityFixture{
			ID:             "uid_soft_deleted",
			UserID:         "user_456",
			Provider:       "github",
			ProviderUserID: "github_user_deleted",
			ProjectID:      "proj_123",
		}
		_, err = testutil.CreateUserIdentity(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err, "Failed to create test user identity")

		// Soft delete the user identity
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE user_identities SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err, "Failed to soft delete user identity")

		// Test FindByProviderAndProviderUserIDAndProjectID - should not find soft-deleted user identity
		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_deleted", "proj_123")
		require.NoError(t, err, "FindByProviderAndProviderUserIDAndProjectID should not return error")
		assert.Nil(t, result, "Result should be nil for soft-deleted user identity")
	})

	t.Run("finds by exact provider, provider_user_id, and project_id combination", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_123",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		// Create multiple user identities with same provider but different provider_user_id
		fixtures := []testutil.UserIdentityFixture{
			{
				ID:             "uid_multi_1",
				UserID:         "user_123",
				Provider:       "github",
				ProviderUserID: "github_user_1",
				ProjectID:      "proj_123",
			},
			{
				ID:             "uid_multi_2",
				UserID:         "user_123",
				Provider:       "github",
				ProviderUserID: "github_user_2",
				ProjectID:      "proj_123",
			},
		}
		_, err = testutil.CreateUserIdentities(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Should find the exact match
		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_1", "proj_123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "uid_multi_1", result.ID)
	})
}

// TestUserIdentityRepository_FindByUserID tests the FindByUserID method
func TestUserIdentityRepository_FindByUserID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserIdentityRepository(testSuite.GetDB())

	t.Run("returns user identities when found", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		userID := "user_123"
		projectID := "proj_123"

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        projectID,
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        userID,
			AccountID: "acc_123",
			ProjectID: projectID,
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixtures := testutil.DefaultUserIdentityFixtures(userID, projectID)
		_, err = testutil.CreateUserIdentities(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err, "Failed to create test user identities")

		result, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err, "FindByUserID should not return error")
		assert.Len(t, result, 3, "Should return all 3 user identities for the user")

		// Verify all returned identities belong to the user
		for _, identity := range result {
			assert.Equal(t, userID, identity.UserID)
		}
	})

	t.Run("returns empty slice when no user identities found", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users")

		result, err := repo.FindByUserID(ctx, "non_existent_user")
		require.NoError(t, err, "FindByUserID should not return error for non-existent user")
		// Result may be nil or empty slice depending on implementation
		if result != nil {
			assert.Empty(t, result, "Result should be empty slice for non-existent user")
		}
	})

	t.Run("excludes soft-deleted user identities", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		userID := "user_456"
		projectID := "proj_123"

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        projectID,
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        userID,
			AccountID: "acc_123",
			ProjectID: projectID,
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixtures := testutil.DefaultUserIdentityFixtures(userID, projectID)
		_, err = testutil.CreateUserIdentities(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Soft delete one of the user identities
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE user_identities SET deleted_at = NOW(4) WHERE id = ?", fixtures[0].ID)
		require.NoError(t, err)

		result, err := repo.FindByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, result, 2, "Should return only 2 user identities (excluding soft-deleted)")

		// Verify the soft-deleted one is not in results
		for _, identity := range result {
			assert.NotEqual(t, fixtures[0].ID, identity.ID)
		}
	})
}

// TestUserIdentityRepository_DeleteByID tests the DeleteByID method
func TestUserIdentityRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserIdentityRepository(testSuite.GetDB())

	t.Run("soft deletes user identity successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_123",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixture := testutil.UserIdentityFixture{
			ID:             "uid_delete_test",
			UserID:         "user_123",
			Provider:       "github",
			ProviderUserID: "github_user_delete",
			ProjectID:      "proj_123",
		}
		_, err = testutil.CreateUserIdentity(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err, "Failed to create test user identity")

		// Delete the user identity
		rowsAffected, err := repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify it's soft deleted (FindByProviderAndProviderUserIDAndProjectID should return nil)
		result, err := repo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_delete", "proj_123")
		require.NoError(t, err)
		assert.Nil(t, result, "Deleted user identity should not be found")

		// Verify it still exists in database with deleted_at set
		var deletedAt *time.Time
		err = testSuite.GetDB().GetContext(ctx, &deletedAt, "SELECT deleted_at FROM user_identities WHERE id = ?", fixture.ID)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt, "deleted_at should be set")
	})

	t.Run("returns zero rows when user identity not found", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users")

		rowsAffected, err := repo.DeleteByID(ctx, "non_existent_id")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should return 0 rows affected for non-existent user identity")
	})

	t.Run("does not delete already deleted user identity", func(t *testing.T) {
		testSuite.CleanTables(t, "user_identities", "users", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_123",
			Name:        "Test Account",
			Description: "Test Description",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create test account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:        "proj_123",
			AccountID: "acc_123",
			Name:      "Test Project",
			IsActive:  true,
		}
		_, err = testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create test project")

		// Create user
		email := "test@example.com"
		userFixture := testutil.UserFixture{
			ID:        "user_789",
			AccountID: "acc_123",
			ProjectID: "proj_123",
			Email:     &email,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err, "Failed to create test user")

		fixture := testutil.UserIdentityFixture{
			ID:             "uid_double_delete",
			UserID:         "user_789",
			Provider:       "github",
			ProviderUserID: "github_user_double",
			ProjectID:      "proj_123",
		}
		_, err = testutil.CreateUserIdentity(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// First delete
		rowsAffected, err := repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Second delete - should return 0 rows
		rowsAffected, err = repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should not delete already deleted user identity")
	})
}
