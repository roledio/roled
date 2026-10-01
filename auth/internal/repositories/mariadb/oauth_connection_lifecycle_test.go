package mariadb

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	"github.com/stretchr/testify/require"
)

func TestOAuthConnectionLifecycleAndTenantIsolation(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	testSuite.CleanTables(t, "oauth_connections", "projects", "accounts")
	_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: "oauth-owner", Name: "Owner", IsActive: true})
	require.NoError(t, err)
	for _, id := range []string{"oauth-project", "oauth-other"} {
		_, err := testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: id, AccountID: "oauth-owner", Name: id, IsActive: true})
		require.NoError(t, err)
	}
	repo := NewOAuthConnectionRepository(db)
	missing, err := repo.FindByProjectIDAndProvider(ctx, "oauth-project", "google")
	require.NoError(t, err)
	require.Nil(t, missing)
	list, err := repo.FindByProjectID(ctx, "oauth-project")
	require.NoError(t, err)
	require.Empty(t, list)
	clientID, secret, scopes := "client", "encrypted-secret", "openid email"
	stored := &entities.OAuthConnection{ID: "oauth-google", ProjectID: "oauth-project", Provider: "google", CredentialType: "custom", ClientID: &clientID, ClientSecretEncrypted: &secret, Scopes: &scopes, Enabled: true}
	require.NoError(t, repo.Create(ctx, stored))
	got, err := repo.FindByProjectIDAndProvider(ctx, stored.ProjectID, stored.Provider)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.False(t, got.CreatedAt.IsZero())
	stored.CreatedAt, stored.UpdatedAt = got.CreatedAt, got.UpdatedAt
	require.Equal(t, stored, got)
	require.Error(t, repo.Create(ctx, stored), "duplicate primary keys must not silently overwrite credentials")
	other := &entities.OAuthConnection{ID: "oauth-other-google", ProjectID: "oauth-other", Provider: "google", CredentialType: "custom", ClientID: &clientID, ClientSecretEncrypted: &secret, Scopes: &scopes, Enabled: true}
	require.NoError(t, repo.Create(ctx, other))
	require.NoError(t, repo.Create(ctx, &entities.OAuthConnection{ID: "oauth-github", ProjectID: stored.ProjectID, Provider: "github", CredentialType: "default", Enabled: true}))
	list, err = repo.FindByProjectID(ctx, stored.ProjectID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, "github", list[0].Provider)
	require.Equal(t, "google", list[1].Provider)

	stored.CredentialType, stored.Enabled = "default", false
	stored.ClientID, stored.ClientSecretEncrypted, stored.Scopes = nil, nil, nil
	count, err := repo.Update(ctx, stored)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	updated, err := repo.FindByProjectIDAndProvider(ctx, stored.ProjectID, stored.Provider)
	require.NoError(t, err)
	stored.UpdatedAt = updated.UpdatedAt
	require.Equal(t, stored, updated)
	unchanged, err := repo.FindByProjectIDAndProvider(ctx, other.ProjectID, other.Provider)
	require.NoError(t, err)
	require.True(t, unchanged.Enabled)
	require.Equal(t, &secret, unchanged.ClientSecretEncrypted)
	count, err = repo.Delete(ctx, stored.ProjectID, stored.Provider)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	missing, err = repo.FindByProjectIDAndProvider(ctx, stored.ProjectID, stored.Provider)
	require.NoError(t, err)
	require.Nil(t, missing)
	unchanged, err = repo.FindByProjectIDAndProvider(ctx, other.ProjectID, other.Provider)
	require.NoError(t, err)
	require.Equal(t, other.ID, unchanged.ID)
}

func TestOAuthConnectionDatabaseFailures(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := NewOAuthConnectionRepository(testSuite.GetDB())
	_, err := repo.FindByProjectID(ctx, "tenant")
	require.ErrorIs(t, err, context.Canceled)
	_, err = repo.FindByProjectIDAndProvider(ctx, "tenant", "google")
	require.ErrorIs(t, err, context.Canceled)
	err = repo.Create(ctx, &entities.OAuthConnection{ID: "connection"})
	require.ErrorIs(t, err, context.Canceled)
	_, err = repo.Update(ctx, &entities.OAuthConnection{ID: "connection"})
	require.ErrorIs(t, err, context.Canceled)
	_, err = repo.Delete(ctx, "tenant", "google")
	require.ErrorIs(t, err, context.Canceled)
}
