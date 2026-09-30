package mariadb

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	"github.com/stretchr/testify/require"
)

func TestBrandingRepositoryUpsertAndIsolation(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	testSuite.CleanTables(t, "brandings", "projects", "accounts")
	_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: "branding-owner", Name: "Branding owner", IsActive: true})
	require.NoError(t, err)
	for _, id := range []string{"branding-project", "other-branding-project"} {
		_, err := testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: id, AccountID: "branding-owner", Name: id, IsActive: true})
		require.NoError(t, err)
	}
	repo := NewBrandingRepository(db)
	missing, err := repo.FindByProjectID(ctx, "branding-project")
	require.NoError(t, err)
	require.Nil(t, missing)
	logo, favicon := "https://example.com/logo.png", "https://example.com/favicon.ico"
	stored := &entities.Branding{ProjectID: "branding-project", LogoURL: &logo, FaviconURL: &favicon, PrimaryColor: "#112233", Rounding: "large", EnableShadow: true, EnableBorder: true}
	affected, err := repo.Upsert(ctx, stored)
	require.NoError(t, err)
	require.Equal(t, 1, affected)
	got, err := repo.FindByProjectID(ctx, stored.ProjectID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.False(t, got.CreatedAt.IsZero())
	require.False(t, got.UpdatedAt.IsZero())
	stored.CreatedAt, stored.UpdatedAt = got.CreatedAt, got.UpdatedAt
	require.Equal(t, stored, got)

	other := &entities.Branding{ProjectID: "other-branding-project", PrimaryColor: "#445566", Rounding: "small", EnableShadow: true}
	_, err = repo.Upsert(ctx, other)
	require.NoError(t, err)
	// Updating an existing project must clear nullable assets and persist false booleans.
	stored.LogoURL, stored.FaviconURL = nil, nil
	stored.EnableShadow, stored.EnableBorder = false, false
	stored.PrimaryColor, stored.Rounding = "#778899", "sharp"
	_, err = repo.Upsert(ctx, stored)
	require.NoError(t, err)
	updated, err := repo.FindByProjectID(ctx, stored.ProjectID)
	require.NoError(t, err)
	require.Equal(t, stored.CreatedAt, updated.CreatedAt)
	require.False(t, updated.UpdatedAt.Before(stored.UpdatedAt))
	stored.UpdatedAt = updated.UpdatedAt
	require.Equal(t, stored, updated)
	unchanged, err := repo.FindByProjectID(ctx, other.ProjectID)
	require.NoError(t, err)
	other.CreatedAt, other.UpdatedAt = unchanged.CreatedAt, unchanged.UpdatedAt
	require.Equal(t, other, unchanged)
	missing, err = repo.FindByProjectID(ctx, "absent-branding-project")
	require.NoError(t, err)
	require.Nil(t, missing)
}

func TestBrandingRepositoryPropagatesDatabaseErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := NewBrandingRepository(testSuite.GetDB())
	_, err := repo.FindByProjectID(ctx, "branding-project")
	require.ErrorIs(t, err, context.Canceled)
	_, err = repo.Upsert(ctx, &entities.Branding{ProjectID: "branding-project"})
	require.ErrorIs(t, err, context.Canceled)
}
