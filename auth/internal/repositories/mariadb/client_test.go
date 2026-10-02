package mariadb

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	pm "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestClientRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	r := NewClientRepository(db)
	testSuite.CleanTables(t, "clients", "projects", "accounts")
	_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: "cr-owner", Name: "Owner", IsActive: true})
	require.NoError(t, err)
	for _, id := range []string{"cr-project", "cr-other"} {
		_, err = testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: id, AccountID: "cr-owner", Name: id, IsActive: true})
		require.NoError(t, err)
	}
	stamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	desc := "special description"
	for i, c := range []entities.Client{
		{ID: "cr-a", AccountID: "cr-owner", ProjectID: "cr-project", Name: "Alpha", Description: &desc, SecretEncrypted: "ciphertext", IsActive: true, IsDefault: true},
		{ID: "cr-b", AccountID: "cr-owner", ProjectID: "cr-project", Name: "Beta", SecretEncrypted: "secret"},
		{ID: "cr-c", AccountID: "cr-owner", ProjectID: "cr-other", Name: "Foreign", SecretEncrypted: "secret", IsActive: true},
	} {
		require.NoError(t, r.Create(ctx, &c))
		_, err = db.ExecContext(ctx, "UPDATE clients SET created_at=? WHERE id=?", stamp.Add(time.Duration(i)*time.Hour), c.ID)
		require.NoError(t, err)
	}
	yes, no := true, false
	since := stamp.Add(time.Hour)
	for _, tc := range []struct {
		name string
		req  models.GetClientsRequest
		want []string
	}{
		{"all", models.GetClientsRequest{}, []string{"cr-a", "cr-b"}},
		{"name", models.GetClientsRequest{Search: " Alpha "}, []string{"cr-a"}},
		{"description", models.GetClientsRequest{Search: "special"}, []string{"cr-a"}},
		{"active", models.GetClientsRequest{IsActive: &yes}, []string{"cr-a"}},
		{"inactive", models.GetClientsRequest{IsActive: &no}, []string{"cr-b"}},
		{"since", models.GetClientsRequest{CreatedAtSince: &since}, []string{"cr-b"}},
		{"until", models.GetClientsRequest{CreatedAtUntil: &stamp}, []string{"cr-a"}},
		{"injection", models.GetClientsRequest{Search: "' OR 1=1 --"}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.ProjectID = "cr-project"
			tc.req.SortBy = "name"
			list, e := r.FindAll(ctx, &tc.req)
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.want, ids)
			n, e := r.Count(ctx, &tc.req)
			require.NoError(t, e)
			require.Equal(t, len(tc.want), n)
		})
	}
	for _, sort := range []string{"name", "is_active", "created_at", "name; DROP TABLE clients"} {
		list, e := r.FindAll(ctx, &models.GetClientsRequest{ProjectID: "cr-project", PageRequest: pm.PageRequest{SortBy: sort, SortDir: "desc", PageSize: 1, PageNum: 2}})
		require.NoError(t, e)
		require.Len(t, list, 1)
	}
	c, err := r.FindByProjectIDAndIsDefault(ctx, "cr-project", true)
	require.NoError(t, err)
	require.Equal(t, "cr-a", c.ID)
	other, err := r.FindByProjectIDAndIsDefault(ctx, "cr-project", false)
	require.NoError(t, err)
	require.Equal(t, "cr-b", other.ID)
	missing, err := r.FindByProjectIDAndIsDefault(ctx, "cr-other", true)
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByIDAndProjectID(ctx, c.ID, "cr-other")
	require.NoError(t, err)
	require.Nil(t, missing)
	c, err = r.FindByIDAndProjectID(ctx, c.ID, "cr-project")
	require.NoError(t, err)
	require.Equal(t, "ciphertext", c.SecretEncrypted)
	c.Name = "Renamed"
	c.Description = nil
	c.IsActive = false
	n, err := r.Update(ctx, c)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	c, err = r.FindByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, "Renamed", c.Name)
	require.Nil(t, c.Description)
	require.False(t, c.IsActive)
	require.Equal(t, "ciphertext", c.SecretEncrypted)
	n, err = r.Delete(ctx, c)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	missing, err = r.FindByID(ctx, c.ID)
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByIDAndProjectID(ctx, c.ID, "cr-project")
	require.NoError(t, err)
	require.Nil(t, missing)
	n, err = r.Update(ctx, c)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = r.Delete(ctx, c)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = r.DeleteByProjectID(ctx, "cr-project")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = r.Count(ctx, &models.GetClientsRequest{ProjectID: "cr-project"})
	require.NoError(t, err)
	require.Zero(t, n)
	foreign, err := r.FindByID(ctx, "cr-c")
	require.NoError(t, err)
	require.NotNil(t, foreign)
	var deleted int
	require.NoError(t, db.GetContext(ctx, &deleted, "SELECT COUNT(*) FROM clients WHERE project_id=? AND deleted_at IS NOT NULL", "cr-project"))
	require.Equal(t, 2, deleted)
}

func TestClientRepositoryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := testSuite.GetDB()
	c := NewClientRepository(db)
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"client create", func() error { return c.Create(ctx, &entities.Client{}) }},
		{"client count", func() error { _, e := c.Count(ctx, &models.GetClientsRequest{}); return e }},
		{"client list", func() error { _, e := c.FindAll(ctx, &models.GetClientsRequest{}); return e }},
		{"client id", func() error { _, e := c.FindByID(ctx, "c"); return e }},
		{"client project", func() error { _, e := c.FindByIDAndProjectID(ctx, "c", "p"); return e }},
		{"client default", func() error { _, e := c.FindByProjectIDAndIsDefault(ctx, "p", true); return e }},
		{"client update", func() error { _, e := c.Update(ctx, &entities.Client{}); return e }},
		{"client delete", func() error { _, e := c.Delete(ctx, &entities.Client{}); return e }},
		{"client delete project", func() error { _, e := c.DeleteByProjectID(ctx, "p"); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) { require.ErrorIs(t, tc.run(), context.Canceled) })
	}
}
