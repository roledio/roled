package mariadb

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessTokenRepository_FindByIDJoin_NullHandling tests NULL value handling in joins
func TestAccessTokenRepository_FindByIDJoin_NullHandling(t *testing.T) {
	ctx := context.Background()
	repo := NewAccessTokenRepository(testSuite.GetDB())

	// Clean all relevant tables before starting tests
	testSuite.CleanTables(t, "access_tokens", "projects", "clients", "users", "user_roles", "roles", "accounts")

	t.Run("handles NULL project_description", func(t *testing.T) {
		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_null_test",
			Name:        "Test Account",
			Description: "Test account for NULL description test",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create account")

		// Create project with NULL description
		projectFixture := testutil.ProjectFixture{
			ID:          "proj_null_desc",
			AccountID:   "acc_null_test",
			Name:        "Project with NULL Description",
			Description: nil, // NULL description
			LogoURL:     nil,
			IsActive:    true,
			IsSystem:    false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create project with NULL description")

		// Create client
		clientFixture := testutil.ClientFixture{
			ID:              "client_null_test",
			AccountID:       "acc_null_test",
			ProjectID:       project.ID,
			Name:            "Test Client",
			Description:     nil,
			SecretEncrypted: "encrypted_secret",
			IsActive:        true,
			IsDefault:       false,
		}
		client, err := testutil.CreateClient(ctx, testSuite.GetDB(), clientFixture)
		require.NoError(t, err, "Failed to create client")

		// Create access token
		tokenFixture := testutil.AccessTokenFixture{
			ID:        "token_null_test",
			AccountID: "acc_null_test",
			ProjectID: project.ID,
			ClientID:  client.ID,
			GrantType: "authorization_code",
			Status:    "issued",
			ExpiresIn: 3600,
			IssuedAt:  time.Now().UTC(),
		}
		token, err := testutil.CreateAccessToken(ctx, testSuite.GetDB(), tokenFixture)
		require.NoError(t, err, "Failed to create access token")

		// Test FindByIDJoin - should handle NULL project_description
		result, err := repo.FindByIDJoin(ctx, token.ID)
		require.NoError(t, err, "FindByIDJoin should not return error with NULL project_description")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions
		assert.Equal(t, token.ID, result.ID, "ID should match")
		assert.Equal(t, project.ID, result.ProjectID, "ProjectID should match")
		assert.Equal(t, "Project with NULL Description", result.ProjectName, "ProjectName should match")
		assert.Nil(t, result.ProjectDescription, "ProjectDescription should be nil for NULL value")
		assert.Nil(t, result.ProjectLogoURL, "ProjectLogoURL should be nil for NULL value")
		assert.Equal(t, client.ID, result.ClientID, "ClientID should match")
		assert.Equal(t, "Test Client", result.ClientName, "ClientName should match")
	})

	t.Run("handles NULL project_logo_url", func(t *testing.T) {
		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_logo_test",
			Name:        "Test Account",
			Description: "Test account for NULL logo test",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create account")

		// Create project with description but NULL logo
		description := "Project with description"
		projectFixture := testutil.ProjectFixture{
			ID:          "proj_null_logo",
			AccountID:   "acc_logo_test",
			Name:        "Project with NULL Logo",
			Description: &description,
			LogoURL:     nil, // NULL logo
			IsActive:    true,
			IsSystem:    false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err, "Failed to create project with NULL logo")

		// Create client
		clientFixture := testutil.ClientFixture{
			ID:              "client_logo_test",
			AccountID:       "acc_logo_test",
			ProjectID:       project.ID,
			Name:            "Test Client",
			Description:     nil,
			SecretEncrypted: "encrypted_secret",
			IsActive:        true,
			IsDefault:       false,
		}
		client, err := testutil.CreateClient(ctx, testSuite.GetDB(), clientFixture)
		require.NoError(t, err, "Failed to create client")

		// Create access token
		tokenFixture := testutil.AccessTokenFixture{
			ID:        "token_logo_test",
			AccountID: "acc_logo_test",
			ProjectID: project.ID,
			ClientID:  client.ID,
			GrantType: "authorization_code",
			Status:    "issued",
			ExpiresIn: 3600,
			IssuedAt:  time.Now().UTC(),
		}
		token, err := testutil.CreateAccessToken(ctx, testSuite.GetDB(), tokenFixture)
		require.NoError(t, err, "Failed to create access token")

		// Test FindByIDJoin - should handle NULL project_logo_url
		result, err := repo.FindByIDJoin(ctx, token.ID)
		require.NoError(t, err, "FindByIDJoin should not return error with NULL project_logo_url")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions
		assert.Equal(t, token.ID, result.ID)
		assert.Equal(t, project.ID, result.ProjectID)
		assert.NotNil(t, result.ProjectDescription, "ProjectDescription should not be nil")
		assert.Equal(t, description, *result.ProjectDescription, "ProjectDescription should match")
		assert.Nil(t, result.ProjectLogoURL, "ProjectLogoURL should be nil for NULL value")
	})

	t.Run("handles NULL user fields", func(t *testing.T) {
		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_user_null",
			Name:        "Test Account",
			Description: "Test account for user NULL test",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:          "proj_user_null",
			AccountID:   "acc_user_null",
			Name:        "Project for User NULL Test",
			Description: nil,
			LogoURL:     nil,
			IsActive:    true,
			IsSystem:    false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create client
		clientFixture := testutil.ClientFixture{
			ID:              "client_user_null",
			AccountID:       "acc_user_null",
			ProjectID:       project.ID,
			Name:            "Test Client",
			Description:     nil,
			SecretEncrypted: "encrypted_secret",
			IsActive:        true,
			IsDefault:       false,
		}
		client, err := testutil.CreateClient(ctx, testSuite.GetDB(), clientFixture)
		require.NoError(t, err)

		// Create user with NULL email and avatar
		userFixture := testutil.UserFixture{
			ID:          "user_null_fields",
			AccountID:   "acc_user_null",
			ProjectID:   project.ID,
			DisplayName: "Test User",
			Email:       nil, // NULL email
			AvatarURL:   nil, // NULL avatar
			ExternalID:  nil, // NULL external ID
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err)

		// Create access token with user
		tokenFixture := testutil.AccessTokenFixture{
			ID:        "token_user_null",
			AccountID: "acc_user_null",
			ProjectID: project.ID,
			ClientID:  client.ID,
			UserID:    &user.ID,
			GrantType: "authorization_code",
			Status:    "issued",
			ExpiresIn: 3600,
			IssuedAt:  time.Now().UTC(),
		}
		token, err := testutil.CreateAccessToken(ctx, testSuite.GetDB(), tokenFixture)
		require.NoError(t, err)

		// Test FindByIDJoin - should handle NULL user fields
		result, err := repo.FindByIDJoin(ctx, token.ID)
		require.NoError(t, err, "FindByIDJoin should not return error with NULL user fields")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions
		assert.NotNil(t, result.UserID, "UserID should not be nil")
		assert.Equal(t, user.ID, *result.UserID, "UserID should match")
		assert.NotNil(t, result.UserDisplayName, "UserDisplayName should not be nil")
		assert.Equal(t, "Test User", *result.UserDisplayName, "UserDisplayName should match")
		assert.Nil(t, result.UserEmail, "UserEmail should be nil for NULL value")
		assert.Nil(t, result.UserAvatarURL, "UserAvatarURL should be nil for NULL value")
		assert.Nil(t, result.UserExternalUserID, "UserExternalUserID should be nil for NULL value")
	})

	t.Run("handles non-NULL values correctly", func(t *testing.T) {
		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_full",
			Name:        "Test Account",
			Description: "Test account for full values test",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create account")

		// Create project with all fields populated
		description := "Full Project Description"
		logoURL := "https://example.com/logo.png"
		projectFixture := testutil.ProjectFixture{
			ID:          "proj_full",
			AccountID:   "acc_full",
			Name:        "Full Project",
			Description: &description,
			LogoURL:     &logoURL,
			IsActive:    true,
			IsSystem:    false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create client
		clientFixture := testutil.ClientFixture{
			ID:              "client_full",
			AccountID:       "acc_full",
			ProjectID:       project.ID,
			Name:            "Full Client",
			Description:     nil,
			SecretEncrypted: "encrypted_secret",
			IsActive:        true,
			IsDefault:       false,
		}
		client, err := testutil.CreateClient(ctx, testSuite.GetDB(), clientFixture)
		require.NoError(t, err)

		// Create access token
		tokenFixture := testutil.AccessTokenFixture{
			ID:        "token_full",
			AccountID: "acc_full",
			ProjectID: project.ID,
			ClientID:  client.ID,
			GrantType: "authorization_code",
			Status:    "issued",
			ExpiresIn: 3600,
			IssuedAt:  time.Now().UTC(),
		}
		token, err := testutil.CreateAccessToken(ctx, testSuite.GetDB(), tokenFixture)
		require.NoError(t, err)

		// Test FindByIDJoin - should handle non-NULL values
		result, err := repo.FindByIDJoin(ctx, token.ID)
		require.NoError(t, err, "FindByIDJoin should not return error with non-NULL values")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions
		assert.NotNil(t, result.ProjectDescription, "ProjectDescription should not be nil")
		assert.Equal(t, description, *result.ProjectDescription, "ProjectDescription should match")
		assert.NotNil(t, result.ProjectLogoURL, "ProjectLogoURL should not be nil")
		assert.Equal(t, logoURL, *result.ProjectLogoURL, "ProjectLogoURL should match")
	})

	t.Run("returns nil for non-existent token", func(t *testing.T) {
		// Test FindByIDJoin with non-existent ID
		result, err := repo.FindByIDJoin(ctx, "non_existent_token_id")
		require.NoError(t, err, "FindByIDJoin should not return error for non-existent token")
		assert.Nil(t, result, "Result should be nil for non-existent token")
	})
}

// TestAccessTokenRepository_FindByIDJoin_WithUserRole tests join with user and role
func TestAccessTokenRepository_FindByIDJoin_WithUserRole(t *testing.T) {
	ctx := context.Background()
	repo := NewAccessTokenRepository(testSuite.GetDB())

	// Clean all relevant tables before starting tests
	testSuite.CleanTables(t, "access_tokens", "projects", "clients", "users", "user_roles", "roles", "accounts")

	t.Run("returns role data when user has role", func(t *testing.T) {
		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_role_test",
			Name:        "Test Account",
			Description: "Test account for role test",
			IsActive:    true,
			IsSystem:    false,
		}
		_, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err, "Failed to create account")

		// Create project
		projectFixture := testutil.ProjectFixture{
			ID:          "proj_role_test",
			AccountID:   "acc_role_test",
			Name:        "Project for Role Test",
			Description: nil,
			LogoURL:     nil,
			IsActive:    true,
			IsSystem:    false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create client
		clientFixture := testutil.ClientFixture{
			ID:              "client_role_test",
			AccountID:       "acc_role_test",
			ProjectID:       project.ID,
			Name:            "Test Client",
			Description:     nil,
			SecretEncrypted: "encrypted_secret",
			IsActive:        true,
			IsDefault:       false,
		}
		client, err := testutil.CreateClient(ctx, testSuite.GetDB(), clientFixture)
		require.NoError(t, err)

		// Create user
		userFixture := testutil.UserFixture{
			ID:          "user_role_test",
			AccountID:   "acc_role_test",
			ProjectID:   project.ID,
			DisplayName: "Test User",
			Email:       nil,
			AvatarURL:   nil,
			ExternalID:  nil,
			IsActive:    true,
		}
		user, err := testutil.CreateUser(ctx, testSuite.GetDB(), userFixture)
		require.NoError(t, err)

		// Create role
		roleFixture := testutil.RoleFixture{
			ID:          "role_test",
			AccountID:   "acc_role_test",
			ProjectID:   project.ID,
			Code:        "admin",
			Name:        "Administrator",
			Description: "Admin role",
		}
		role, err := testutil.CreateRole(ctx, testSuite.GetDB(), roleFixture)
		require.NoError(t, err)

		// Assign role to user
		_, err = testSuite.GetDB().ExecContext(ctx,
			"INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, NOW(4))",
			user.ID, role.ID)
		require.NoError(t, err)

		// Create access token with user
		tokenFixture := testutil.AccessTokenFixture{
			ID:        "token_role_test",
			AccountID: "acc_role_test",
			ProjectID: project.ID,
			ClientID:  client.ID,
			UserID:    &user.ID,
			GrantType: "authorization_code",
			Status:    "issued",
			ExpiresIn: 3600,
			IssuedAt:  time.Now().UTC(),
		}
		token, err := testutil.CreateAccessToken(ctx, testSuite.GetDB(), tokenFixture)
		require.NoError(t, err)

		// Test FindByIDJoin - should return role data
		result, err := repo.FindByIDJoin(ctx, token.ID)
		require.NoError(t, err, "FindByIDJoin should not return error with role data")
		require.NotNil(t, result, "Result should not be nil")

		// Assertions for role data
		assert.NotNil(t, result.RoleID, "RoleID should not be nil")
		assert.Equal(t, role.ID, *result.RoleID, "RoleID should match")
		assert.NotNil(t, result.RoleCode, "RoleCode should not be nil")
		assert.Equal(t, "admin", *result.RoleCode, "RoleCode should match")
		assert.NotNil(t, result.RoleName, "RoleName should not be nil")
		assert.Equal(t, "Administrator", *result.RoleName, "RoleName should match")
		assert.NotNil(t, result.RoleDescription, "RoleDescription should not be nil")
		assert.Equal(t, "Admin role", *result.RoleDescription, "RoleDescription should match")
		assert.NotNil(t, result.UserDisplayName, "UserDisplayName should not be nil")
		assert.Equal(t, "Test User", *result.UserDisplayName, "UserDisplayName should match")
	})
}
