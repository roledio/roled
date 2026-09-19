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

// TestOAuthConnectionRepository_FindByProjectID tests the FindByProjectID method
func TestOAuthConnectionRepository_FindByProjectID(t *testing.T) {
	ctx := context.Background()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())

	t.Run("returns connections when found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_test",
			Name:        "OAuth Test Account",
			Description: "Test account for OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_test",
			AccountID: account.ID,
			Name:      "OAuth Test Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connections
		fixtures := testutil.DefaultOAuthConnectionFixtures(project.ID)
		connections, err := testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)
		require.Len(t, connections, 3)

		// Test FindByProjectID
		result, err := repo.FindByProjectID(ctx, project.ID)
		require.NoError(t, err, "FindByProjectID should not return error")
		require.Len(t, result, 3, "Should return all 3 connections")

		// Verify ordering (should be ordered by provider ASC)
		assert.Equal(t, "custom_provider", result[0].Provider)
		assert.Equal(t, "github", result[1].Provider)
		assert.Equal(t, "google", result[2].Provider)
	})

	t.Run("returns empty slice when no connections found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_empty",
			Name:        "OAuth Empty Account",
			Description: "Test account for empty OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_empty",
			AccountID: account.ID,
			Name:      "OAuth Empty Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Test FindByProjectID with no connections
		result, err := repo.FindByProjectID(ctx, project.ID)
		require.NoError(t, err, "FindByProjectID should not return error")
		assert.Len(t, result, 0, "Should return empty slice")
	})

	t.Run("returns connections for specific project only", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_multi",
			Name:        "OAuth Multi Account",
			Description: "Test account for multiple projects",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		// Create two projects
		project1Fixture := testutil.ProjectFixture{
			ID:        "proj_oauth_1",
			AccountID: account.ID,
			Name:      "OAuth Project 1",
			IsActive:  true,
			IsSystem:  false,
		}
		project1, err := testutil.CreateProject(ctx, testSuite.GetDB(), project1Fixture)
		require.NoError(t, err)

		project2Fixture := testutil.ProjectFixture{
			ID:        "proj_oauth_2",
			AccountID: account.ID,
			Name:      "OAuth Project 2",
			IsActive:  true,
			IsSystem:  false,
		}
		project2, err := testutil.CreateProject(ctx, testSuite.GetDB(), project2Fixture)
		require.NoError(t, err)

		// Create connections for project 1 with unique IDs
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		scopes := "read,write"

		fixtures1 := []testutil.OAuthConnectionFixture{
			{
				ID:                    "oauth_p1_github",
				ProjectID:             project1.ID,
				Provider:              "github",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
			{
				ID:                    "oauth_p1_google",
				ProjectID:             project1.ID,
				Provider:              "google",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
		}
		_, err = testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures1)
		require.NoError(t, err)

		// Create connections for project 2 with unique IDs
		fixtures2 := []testutil.OAuthConnectionFixture{
			{
				ID:                    "oauth_p2_github",
				ProjectID:             project2.ID,
				Provider:              "github",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
			{
				ID:                    "oauth_p2_google",
				ProjectID:             project2.ID,
				Provider:              "google",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
		}
		_, err = testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures2)
		require.NoError(t, err)

		// Test FindByProjectID for project 1
		result1, err := repo.FindByProjectID(ctx, project1.ID)
		require.NoError(t, err)
		assert.Len(t, result1, 2, "Should return 2 connections for project 1")

		// Test FindByProjectID for project 2
		result2, err := repo.FindByProjectID(ctx, project2.ID)
		require.NoError(t, err)
		assert.Len(t, result2, 2, "Should return 2 connections for project 2")

		// Verify they're different connections
		assert.NotEqual(t, result1[0].ID, result2[0].ID)
	})
}

// TestOAuthConnectionRepository_FindByProjectIDAndProvider tests the FindByProjectIDAndProvider method
func TestOAuthConnectionRepository_FindByProjectIDAndProvider(t *testing.T) {
	ctx := context.Background()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())

	t.Run("returns connection when found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_find",
			Name:        "OAuth Find Account",
			Description: "Test account for find OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_find",
			AccountID: account.ID,
			Name:      "OAuth Find Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connection
		fixtures := testutil.DefaultOAuthConnectionFixtures(project.ID)
		connections, err := testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Test FindByProjectIDAndProvider
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "github")
		require.NoError(t, err, "FindByProjectIDAndProvider should not return error")
		require.NotNil(t, result, "Result should not be nil")

		// Find the github connection from fixtures by provider
		var githubConnection *entities.OAuthConnection
		for i := range connections {
			if connections[i].Provider == "github" {
				githubConnection = &connections[i]
				break
			}
		}
		require.NotNil(t, githubConnection, "GitHub connection should exist in fixtures")

		// Verify the connection
		assert.Equal(t, githubConnection.ID, result.ID)
		assert.Equal(t, project.ID, result.ProjectID)
		assert.Equal(t, "github", result.Provider)
		assert.True(t, result.Enabled)
	})

	t.Run("returns nil when connection not found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_notfound",
			Name:        "OAuth Not Found Account",
			Description: "Test account for not found OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_notfound",
			AccountID: account.ID,
			Name:      "OAuth Not Found Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Test FindByProjectIDAndProvider with non-existent provider
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "nonexistent")
		require.NoError(t, err, "FindByProjectIDAndProvider should not return error for non-existent provider")
		assert.Nil(t, result, "Result should be nil for non-existent provider")
	})

	t.Run("returns nil when project not found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections")

		// Test FindByProjectIDAndProvider with non-existent project
		result, err := repo.FindByProjectIDAndProvider(ctx, "nonexistent_project", "github")
		require.NoError(t, err, "FindByProjectIDAndProvider should not return error for non-existent project")
		assert.Nil(t, result, "Result should be nil for non-existent project")
	})
}

// TestOAuthConnectionRepository_Create tests the Create method
func TestOAuthConnectionRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())

	t.Run("creates connection successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_create",
			Name:        "OAuth Create Account",
			Description: "Test account for create OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_create",
			AccountID: account.ID,
			Name:      "OAuth Create Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connection
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		scopes := "read,write"

		connection := &entities.OAuthConnection{
			ID:                    "oauth_create_test",
			ProjectID:             project.ID,
			Provider:              "github",
			CredentialType:        "default",
			ClientID:              &clientID,
			ClientSecretEncrypted: &clientSecret,
			Scopes:                &scopes,
			Enabled:               true,
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err = repo.Create(ctx, connection)
		require.NoError(t, err, "Create should not return error")

		// Verify the connection was created
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "github")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, connection.ID, result.ID)
		assert.Equal(t, connection.ProjectID, result.ProjectID)
		assert.Equal(t, connection.Provider, result.Provider)
		assert.Equal(t, connection.CredentialType, result.CredentialType)
		assert.Equal(t, connection.ClientID, result.ClientID)
		assert.Equal(t, connection.ClientSecretEncrypted, result.ClientSecretEncrypted)
		assert.Equal(t, connection.Scopes, result.Scopes)
		assert.Equal(t, connection.Enabled, result.Enabled)
	})

	t.Run("creates connection with nullable fields", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_nullable",
			Name:        "OAuth Nullable Account",
			Description: "Test account for nullable OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_nullable",
			AccountID: account.ID,
			Name:      "OAuth Nullable Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connection with nullable fields
		connection := &entities.OAuthConnection{
			ID:                    "oauth_nullable_test",
			ProjectID:             project.ID,
			Provider:              "custom_provider",
			CredentialType:        "custom",
			ClientID:              nil,
			ClientSecretEncrypted: nil,
			Scopes:                nil,
			Enabled:               false,
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err = repo.Create(ctx, connection)
		require.NoError(t, err, "Create should not return error")

		// Verify the connection was created with nil fields
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "custom_provider")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.ClientID)
		assert.Nil(t, result.ClientSecretEncrypted)
		assert.Nil(t, result.Scopes)
		assert.False(t, result.Enabled)
	})
}

// TestOAuthConnectionRepository_Update tests the Update method
func TestOAuthConnectionRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())

	t.Run("updates connection successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_update",
			Name:        "OAuth Update Account",
			Description: "Test account for update OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_update",
			AccountID: account.ID,
			Name:      "OAuth Update Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connection
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		scopes := "read,write"

		connection := &entities.OAuthConnection{
			ID:                    "oauth_update_test",
			ProjectID:             project.ID,
			Provider:              "github",
			CredentialType:        "default",
			ClientID:              &clientID,
			ClientSecretEncrypted: &clientSecret,
			Scopes:                &scopes,
			Enabled:               true,
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err = repo.Create(ctx, connection)
		require.NoError(t, err)

		// Update the connection
		newClientID := "new_client_id"
		newClientSecret := "new_encrypted_secret"
		newScopes := "read,write,admin"

		updatedConnection := &entities.OAuthConnection{
			ID:                    connection.ID,
			ProjectID:             project.ID,
			Provider:              "github",
			CredentialType:        "custom",
			ClientID:              &newClientID,
			ClientSecretEncrypted: &newClientSecret,
			Scopes:                &newScopes,
			Enabled:               false,
			UpdatedAt:             time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedConnection)
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify the update
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "github")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, newClientID, *result.ClientID)
		assert.Equal(t, newClientSecret, *result.ClientSecretEncrypted)
		assert.Equal(t, newScopes, *result.Scopes)
		assert.Equal(t, "custom", result.CredentialType)
		assert.False(t, result.Enabled)
	})

	t.Run("returns zero rows when connection not found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_nu",
			Name:        "OAuth Not Found Update Account",
			Description: "Test account for not found update OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_nu",
			AccountID: account.ID,
			Name:      "OAuth Not Found Update Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Try to update non-existent connection
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		updatedConnection := &entities.OAuthConnection{
			ID:                    "nonexistent_conn",
			ProjectID:             project.ID,
			Provider:              "github",
			CredentialType:        "default",
			ClientID:              &clientID,
			ClientSecretEncrypted: &clientSecret,
			Enabled:               true,
			UpdatedAt:             time.Now().UTC(),
		}

		rowsAffected, err := repo.Update(ctx, updatedConnection)
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should return 0 rows affected for non-existent connection")
	})
}

// TestOAuthConnectionRepository_Delete tests the Delete method
func TestOAuthConnectionRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())

	t.Run("deletes connection successfully", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_delete",
			Name:        "OAuth Delete Account",
			Description: "Test account for delete OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_delete",
			AccountID: account.ID,
			Name:      "OAuth Delete Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connections with unique IDs
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		scopes := "read,write"

		fixtures := []testutil.OAuthConnectionFixture{
			{
				ID:                    "oauth_del_github",
				ProjectID:             project.ID,
				Provider:              "github",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
			{
				ID:                    "oauth_del_google",
				ProjectID:             project.ID,
				Provider:              "google",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
			{
				ID:                    "oauth_del_custom",
				ProjectID:             project.ID,
				Provider:              "custom_provider",
				CredentialType:        "custom",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               false,
			},
		}
		_, err = testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Delete the connection
		rowsAffected, err := repo.Delete(ctx, project.ID, "github")
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify it's deleted
		result, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "github")
		require.NoError(t, err)
		assert.Nil(t, result, "Deleted connection should not be found")

		// Verify other connections still exist
		remaining, err := repo.FindByProjectID(ctx, project.ID)
		require.NoError(t, err)
		assert.Len(t, remaining, 2, "Should have 2 remaining connections")
	})

	t.Run("returns zero rows when connection not found", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_nd",
			Name:        "OAuth Not Found Delete Account",
			Description: "Test account for not found delete OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_nd",
			AccountID: account.ID,
			Name:      "OAuth Not Found Delete Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Try to delete non-existent connection
		rowsAffected, err := repo.Delete(ctx, project.ID, "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, 0, rowsAffected, "Should return 0 rows affected for non-existent connection")
	})

	t.Run("deletes connection by project_id and provider only", func(t *testing.T) {
		testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")

		// Create account and project
		accountFixture := testutil.AccountFixture{
			ID:          "acc_oauth_sd",
			Name:        "OAuth Specific Delete Account",
			Description: "Test account for specific delete OAuth",
			IsActive:    true,
			IsSystem:    false,
		}
		account, err := testutil.CreateAccount(ctx, testSuite.GetDB(), accountFixture)
		require.NoError(t, err)

		projectFixture := testutil.ProjectFixture{
			ID:        "proj_oauth_sd",
			AccountID: account.ID,
			Name:      "OAuth Specific Delete Project",
			IsActive:  true,
			IsSystem:  false,
		}
		project, err := testutil.CreateProject(ctx, testSuite.GetDB(), projectFixture)
		require.NoError(t, err)

		// Create OAuth connections with unique IDs
		clientID := "test_client_id"
		clientSecret := "encrypted_secret"
		scopes := "read,write"

		fixtures := []testutil.OAuthConnectionFixture{
			{
				ID:                    "oauth_sd_github",
				ProjectID:             project.ID,
				Provider:              "github",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
			{
				ID:                    "oauth_sd_google",
				ProjectID:             project.ID,
				Provider:              "google",
				CredentialType:        "default",
				ClientID:              &clientID,
				ClientSecretEncrypted: &clientSecret,
				Scopes:                &scopes,
				Enabled:               true,
			},
		}
		_, err = testutil.CreateOAuthConnections(ctx, testSuite.GetDB(), fixtures)
		require.NoError(t, err)

		// Delete specific connection by provider
		rowsAffected, err := repo.Delete(ctx, project.ID, "github")
		require.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)

		// Verify only github connection is deleted
		githubResult, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "github")
		require.NoError(t, err)
		assert.Nil(t, githubResult, "GitHub connection should be deleted")

		googleResult, err := repo.FindByProjectIDAndProvider(ctx, project.ID, "google")
		require.NoError(t, err)
		assert.NotNil(t, googleResult, "Google connection should still exist")
	})
}
