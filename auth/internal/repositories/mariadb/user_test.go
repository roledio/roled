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
	"go.openly.dev/pointy"
)

// TestUserRepository_FindByID tests the FindByID method
func TestUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		// Setup account and project
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_find_by_id",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_find_by_id",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "test@example.com"
		externalID := "ext-123"
		fixture := testutil.UserFixture{
			ID:          "user_find_by_id",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Test User",
			Email:       &email,
			ExternalID:  &externalID,
			IsActive:    true,
		}
		expected, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Test FindByID
		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, expected.ID, result.ID)
		assert.Equal(t, expected.DisplayName, result.DisplayName)
		assert.Equal(t, expected.Email, result.Email)
		assert.Equal(t, expected.ExternalUserID, result.ExternalUserID)
		assert.Equal(t, expected.IsActive, result.IsActive)
		assert.Nil(t, result.DeletedAt)
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users")

		result, err := repo.FindByID(ctx, "non_existent_id")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("excludes soft-deleted users", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_soft_del",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_soft_del",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_soft_del",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Soft Deleted User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Soft delete the user
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestUserRepository_FindByProjectIDAndEmail tests the FindByProjectIDAndEmail method
func TestUserRepository_FindByProjectIDAndEmail(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_email",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_email",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "test@example.com"
		fixture := testutil.UserFixture{
			ID:          "user_email",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Test User",
			Email:       &email,
			IsActive:    true,
		}
		expected, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndEmail(ctx, project.ID, email)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, expected.ID, result.ID)
		assert.Equal(t, expected.Email, result.Email)
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_not_found",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_not_found",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndEmail(ctx, project.ID, "nonexistent@example.com")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("excludes soft-deleted users", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_soft_del_email",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_soft_del_email",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "softdel@example.com"
		fixture := testutil.UserFixture{
			ID:          "user_soft_del_email",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Soft Deleted User",
			Email:       &email,
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndEmail(ctx, project.ID, email)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestUserRepository_FindByProjectIDAndExternalUserID tests the FindByProjectIDAndExternalUserID method
func TestUserRepository_FindByProjectIDAndExternalUserID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_ext",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_ext",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		externalID := "google_12345"
		fixture := testutil.UserFixture{
			ID:          "user_ext",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "External User",
			ExternalID:  &externalID,
			IsActive:    true,
		}
		expected, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndExternalUserID(ctx, project.ID, externalID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, expected.ID, result.ID)
		assert.Equal(t, expected.ExternalUserID, result.ExternalUserID)
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_ext_not",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_ext_not",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndExternalUserID(ctx, project.ID, "nonexistent_ext_id")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestUserRepository_FindByProjectIDAndExternalUserIDJoinRole tests the join method
func TestUserRepository_FindByProjectIDAndExternalUserIDJoinRole(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user with role when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects", "roles", "user_roles")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_join",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_join",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		externalID := "google_join_123"
		fixture := testutil.UserFixture{
			ID:          "user_join",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Join User",
			ExternalID:  &externalID,
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Create role and assign to user
		role, err := testutil.CreateRole(ctx, testSuite.GetDB(), testutil.RoleFixture{
			ID:          "role_join",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Code:        "admin",
			Name:        "Admin",
			Description: "Administrator role",
		})
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndExternalUserIDJoinRole(ctx, project.ID, externalID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, role.ID, result.RoleID)
		assert.Equal(t, role.Name, result.RoleName)
	})

	t.Run("returns user without role when not assigned", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_no_role",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_no_role",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		externalID := "google_no_role_123"
		fixture := testutil.UserFixture{
			ID:          "user_no_role",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "No Role User",
			ExternalID:  &externalID,
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		result, err := repo.FindByProjectIDAndExternalUserIDJoinRole(ctx, project.ID, externalID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, user.ID, result.ID)
		assert.Empty(t, result.RoleID)
		assert.Empty(t, result.RoleName)
	})
}

// TestUserRepository_FindByIDAndProjectID tests the FindByIDAndProjectID method
func TestUserRepository_FindByIDAndProjectID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_id_proj",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_id_proj",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_id_proj",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "ID Project User",
			IsActive:    true,
		}
		expected, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		result, err := repo.FindByIDAndProjectID(ctx, fixture.ID, project.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, expected.ID, result.ID)
		assert.Equal(t, expected.ProjectID, result.ProjectID)
	})

	t.Run("returns nil when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_id_proj_not",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_id_proj_not",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		result, err := repo.FindByIDAndProjectID(ctx, "nonexistent_id", project.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns nil when user exists but in different project", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_diff_proj",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project1, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_diff_proj_1",
			AccountID: account.ID,
			Name:      "Test Project 1",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		project2, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_diff_proj_2",
			AccountID: account.ID,
			Name:      "Test Project 2",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_diff_proj",
			AccountID:   account.ID,
			ProjectID:   project1.ID,
			DisplayName: "Different Project User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		result, err := repo.FindByIDAndProjectID(ctx, fixture.ID, project2.ID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestUserRepository_FindByIDAndProjectIDJoinRole tests the join method
func TestUserRepository_FindByIDAndProjectIDJoinRole(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns user with role when found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects", "roles", "user_roles")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_id_join",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_id_join",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_id_join",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "ID Join User",
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		role, err := testutil.CreateRole(ctx, testSuite.GetDB(), testutil.RoleFixture{
			ID:          "role_id_join",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Code:        "user",
			Name:        "User",
			Description: "User role",
		})
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)
		require.NoError(t, err)

		result, err := repo.FindByIDAndProjectIDJoinRole(ctx, user.ID, project.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, role.ID, result.RoleID)
		assert.Equal(t, role.Name, result.RoleName)
	})
}

// TestUserRepository_Create tests the Create method
func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("creates user successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_create",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_create",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "newuser@example.com"
		externalID := "ext_new_123"
		user := &entities.User{
			ID:             "user_create_test",
			AccountID:      account.ID,
			ProjectID:      project.ID,
			DisplayName:    "New User",
			Email:          &email,
			ExternalUserID: &externalID,
			AvatarURL:      pointy.String("https://example.com/avatar.png"),
			IsActive:       true,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)

		result, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, user.DisplayName, result.DisplayName)
		assert.Equal(t, user.Email, result.Email)
		assert.Equal(t, user.ExternalUserID, result.ExternalUserID)
		assert.Equal(t, user.AvatarURL, result.AvatarURL)
		assert.Equal(t, user.IsActive, result.IsActive)
	})

	t.Run("creates user with password hash", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_create_pass",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_create_pass",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "passworduser@example.com"
		passwordHash := "hashed_password_123"
		user := &entities.User{
			ID:           "user_create_pass",
			AccountID:    account.ID,
			ProjectID:    project.ID,
			DisplayName:  "Password User",
			Email:        &email,
			PasswordHash: &passwordHash,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)

		result, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, passwordHash, *result.PasswordHash)
	})
}

// TestUserRepository_Update tests the Update method
func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("updates user successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_update",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_update",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "old@example.com"
		fixture := testutil.UserFixture{
			ID:          "user_update",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Old Name",
			Email:       &email,
			IsActive:    false,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		newEmail := "new@example.com"
		updatedUser := &entities.User{
			ID:          fixture.ID,
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Updated Name",
			Email:       &newEmail,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedUser)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Updated Name", result.DisplayName)
		assert.Equal(t, newEmail, *result.Email)
		assert.True(t, result.IsActive)
	})

	t.Run("returns zero rows when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_update_not",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_update_not",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email := "notfound@example.com"
		updatedUser := &entities.User{
			ID:          "nonexistent_id",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Nonexistent",
			Email:       &email,
			IsActive:    true,
			UpdatedAt:   time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedUser)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
	})
}

// TestUserRepository_SetEmailVerified tests the SetEmailVerified method
func TestUserRepository_SetEmailVerified(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("sets email verified successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_verify",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_verify",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_verify",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Unverified User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		rowsAffected, err := repo.SetEmailVerified(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotNil(t, result.EmailVerifiedAt)
	})

	t.Run("returns zero rows when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users")

		rowsAffected, err := repo.SetEmailVerified(ctx, "nonexistent_id")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
	})

	t.Run("does not update already verified user", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_verify_already",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_verify_already",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		now := time.Now().UTC()
		fixture := testutil.UserFixture{
			ID:          "user_verify_already",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Already Verified User",
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Set email verified
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET email_verified_at = ? WHERE id = ?", now, user.ID)
		require.NoError(t, err)

		// Try to set again
		rowsAffected, err := repo.SetEmailVerified(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
	})
}

// TestUserRepository_UpdatePassword tests the UpdatePassword method
func TestUserRepository_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("updates password successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_pass",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_pass",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		passwordHash := "old_hash_123"
		fixture := testutil.UserFixture{
			ID:          "user_pass",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Password User",
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		// Set initial password
		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, user.ID)
		require.NoError(t, err)

		newPasswordHash := "new_hash_456"
		rowsAffected, err := repo.UpdatePassword(ctx, user.ID, newPasswordHash)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		result, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, newPasswordHash, *result.PasswordHash)
	})

	t.Run("returns zero rows when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users")

		rowsAffected, err := repo.UpdatePassword(ctx, "nonexistent_id", "hash123")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
	})
}

// TestUserRepository_DeleteByID tests the DeleteByID method
func TestUserRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("soft deletes user successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_delete",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_delete",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_delete",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "User to Delete",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		rowsAffected, err := repo.DeleteByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		result, err := repo.FindByID(ctx, fixture.ID)
		require.NoError(t, err)
		assert.Nil(t, result)

		var deletedAt *time.Time
		err = testSuite.GetDB().GetContext(ctx, &deletedAt, "SELECT deleted_at FROM users WHERE id = ?", fixture.ID)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("returns zero rows when user not found", func(t *testing.T) {
		testSuite.CleanTables(t, "users")

		rowsAffected, err := repo.DeleteByID(ctx, "nonexistent_id")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
	})
}

// TestUserRepository_DeleteByAccountID tests the DeleteByAccountID method
func TestUserRepository_DeleteByAccountID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("soft deletes users by account ID", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_del_acc",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_del_acc",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create multiple users
		for i := 0; i < 3; i++ {
			fixture := testutil.UserFixture{
				ID:          fmt.Sprintf("user_del_acc_%d", i),
				AccountID:   account.ID,
				ProjectID:   project.ID,
				DisplayName: fmt.Sprintf("User %d", i),
				IsActive:    true,
			}
			_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		rowsAffected, err := repo.DeleteByAccountID(ctx, account.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, rowsAffected)
	})
}

// TestUserRepository_DeleteByProjectID tests the DeleteByProjectID method
func TestUserRepository_DeleteByProjectID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("soft deletes users by project ID", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_del_proj",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_del_proj",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create multiple users
		for i := 0; i < 3; i++ {
			fixture := testutil.UserFixture{
				ID:          fmt.Sprintf("user_del_proj_%d", i),
				AccountID:   account.ID,
				ProjectID:   project.ID,
				DisplayName: fmt.Sprintf("User %d", i),
				IsActive:    true,
			}
			_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		rowsAffected, err := repo.DeleteByProjectID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, rowsAffected)
	})
}

// TestUserRepository_Count tests the Count method
func TestUserRepository_Count(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("counts all users", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_count",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_count",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create multiple users
		for i := 0; i < 5; i++ {
			fixture := testutil.UserFixture{
				ID:          fmt.Sprintf("user_count_%d", i),
				AccountID:   account.ID,
				ProjectID:   project.ID,
				DisplayName: fmt.Sprintf("User %d", i),
				IsActive:    true,
			}
			_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		req := &models.GetUsersRequest{ProjectID: project.ID}
		count, err := repo.Count(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("filters by is_active", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_count_active",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_count_active",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create active and inactive users
		isActive := true
		fixture := testutil.UserFixture{
			ID:          "user_count_active_1",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Active User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		fixtureInactive := testutil.UserFixture{
			ID:          "user_count_active_2",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Inactive User",
			IsActive:    false,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixtureInactive)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID, IsActive: &isActive}
		count, err := repo.Count(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("searches by email or display name", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_count_search",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_count_search",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		email1 := "search@example.com"
		fixture1 := testutil.UserFixture{
			ID:          "user_count_search_1",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Search User",
			Email:       &email1,
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture1)
		require.NoError(t, err)

		email2 := "other@example.com"
		fixture2 := testutil.UserFixture{
			ID:          "user_count_search_2",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Other User",
			Email:       &email2,
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture2)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID, Search: "search"}
		count, err := repo.Count(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("filters by role_id", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects", "roles", "user_roles")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_count_role",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_count_role",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create users and role
		role, err := testutil.CreateRole(ctx, testSuite.GetDB(), testutil.RoleFixture{
			ID:          "role_count",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Code:        "admin",
			Name:        "Admin",
			Description: "Admin role",
		})
		require.NoError(t, err)

		fixture1 := testutil.UserFixture{
			ID:          "user_count_role_1",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Admin User",
			IsActive:    true,
		}
		user1, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture1)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user1.ID, role.ID)
		require.NoError(t, err)

		fixture2 := testutil.UserFixture{
			ID:          "user_count_role_2",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Regular User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture2)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID, RoleID: &role.ID}
		count, err := repo.Count(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("excludes soft-deleted users from count", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_count_del",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_count_del",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_count_del",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Deleted User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID}
		count, err := repo.Count(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestUserRepository_FindAll tests the FindAll method
func TestUserRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testSuite.GetDB())

	t.Run("returns all users with default pagination", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_findall",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_findall",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create multiple users
		for i := 0; i < 5; i++ {
			fixture := testutil.UserFixture{
				ID:          fmt.Sprintf("user_findall_%d", i),
				AccountID:   account.ID,
				ProjectID:   project.ID,
				DisplayName: fmt.Sprintf("User %d", i),
				IsActive:    true,
			}
			_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		req := &models.GetUsersRequest{ProjectID: project.ID}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 5)
	})

	t.Run("supports pagination", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_page",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_page",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		// Create more users for pagination
		for i := 0; i < 15; i++ {
			fixture := testutil.UserFixture{
				ID:          fmt.Sprintf("user_page_%02d", i),
				AccountID:   account.ID,
				ProjectID:   project.ID,
				DisplayName: fmt.Sprintf("User %d", i),
				IsActive:    true,
			}
			_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
			require.NoError(t, err)
		}

		req := &models.GetUsersRequest{
			ProjectID: project.ID,
			PageRequest: pkgmodels.PageRequest{
				PageNum:  1,
				PageSize: 5,
			},
		}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 5)

		// Test second page
		req.PageNum = 2
		result2, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result2, 5)

		assert.NotEqual(t, result[0].ID, result2[0].ID)
	})

	t.Run("supports sorting", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_sort",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_sort",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture1 := testutil.UserFixture{
			ID:          "user_sort_1",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Zebra User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture1)
		require.NoError(t, err)

		fixture2 := testutil.UserFixture{
			ID:          "user_sort_2",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Alpha User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture2)
		require.NoError(t, err)

		req := &models.GetUsersRequest{
			ProjectID: project.ID,
			PageRequest: pkgmodels.PageRequest{
				PageNum:  1,
				PageSize: 10,
				SortBy:   "display_name",
				SortDir:  "ASC",
			},
		}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Alpha User", result[0].DisplayName)
		assert.Equal(t, "Zebra User", result[1].DisplayName)
	})

	t.Run("filters by is_active", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_filter_active",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_filter_active",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		isActive := true
		fixture1 := testutil.UserFixture{
			ID:          "user_filter_active_1",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Active User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture1)
		require.NoError(t, err)

		fixture2 := testutil.UserFixture{
			ID:          "user_filter_active_2",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Inactive User",
			IsActive:    false,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture2)
		require.NoError(t, err)

		req := &models.GetUsersRequest{
			ProjectID: project.ID,
			IsActive:  &isActive,
		}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.True(t, result[0].IsActive)
	})

	t.Run("includes role information", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects", "roles", "user_roles")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_join_role",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_join_role",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		role, err := testutil.CreateRole(ctx, testSuite.GetDB(), testutil.RoleFixture{
			ID:          "role_join",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Code:        "admin",
			Name:        "Admin",
			Description: "Admin role",
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_join_role",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Admin User",
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, role.ID, result[0].RoleID)
		assert.Equal(t, role.Name, result[0].RoleName)
	})

	t.Run("excludes soft-deleted users", func(t *testing.T) {
		testSuite.CleanTables(t, "users", "accounts", "projects")

		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), testutil.AccountFixture{
			ID:       "acc_findall_del",
			Name:     "Test Account",
			IsActive: true,
			IsSystem: false,
		})
		require.NoError(t, err)

		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), testutil.ProjectFixture{
			ID:        "proj_findall_del",
			AccountID: account.ID,
			Name:      "Test Project",
			IsActive:  true,
			IsSystem:  false,
		})
		require.NoError(t, err)

		fixture := testutil.UserFixture{
			ID:          "user_findall_del",
			AccountID:   account.ID,
			ProjectID:   project.ID,
			DisplayName: "Deleted User",
			IsActive:    true,
		}
		_, err = testutil.CreateUser(ctx, testSuite.GetDB(), fixture)
		require.NoError(t, err)

		_, err = testSuite.GetDB().ExecContext(ctx, "UPDATE users SET deleted_at = NOW(4) WHERE id = ?", fixture.ID)
		require.NoError(t, err)

		req := &models.GetUsersRequest{ProjectID: project.ID}
		req.SetDefaults()

		result, err := repo.FindAll(ctx, req)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}
