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

func TestProjectRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	r := NewProjectRepository(db)
	testSuite.CleanTables(t, "projects", "accounts")
	for _, id := range []string{"pr-owner", "pr-other"} {
		_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: id, Name: id, IsActive: true})
		require.NoError(t, err)
	}
	missing, err := r.FindSystem(ctx)
	require.NoError(t, err)
	require.Nil(t, missing)
	stamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	desc := "special description"
	logo := "https://example.com/logo.png"
	for i, p := range []entities.Project{
		{ID: "pr-a", AccountID: "pr-owner", Name: "Alpha", Description: &desc, LogoURL: &logo, IsActive: true, IsSystem: true},
		{ID: "pr-b", AccountID: "pr-owner", Name: "Beta"},
		{ID: "pr-c", AccountID: "pr-other", Name: "Foreign", IsActive: true},
	} {
		require.NoError(t, r.Create(ctx, &p))
		_, err = db.ExecContext(ctx, "UPDATE projects SET created_at=? WHERE id=?", stamp.Add(time.Duration(i)*time.Hour), p.ID)
		require.NoError(t, err)
	}
	yes, no := true, false
	since := stamp.Add(time.Hour)
	for _, tc := range []struct {
		name string
		req  models.GetProjectsRequest
		want []string
	}{
		{"all", models.GetProjectsRequest{}, []string{"pr-a", "pr-b"}},
		{"name", models.GetProjectsRequest{Search: " Alpha "}, []string{"pr-a"}},
		{"description", models.GetProjectsRequest{Search: "special"}, []string{"pr-a"}},
		{"active", models.GetProjectsRequest{IsActive: &yes}, []string{"pr-a"}},
		{"inactive", models.GetProjectsRequest{IsActive: &no}, []string{"pr-b"}},
		{"since inclusive", models.GetProjectsRequest{CreatedAtSince: &since}, []string{"pr-b"}},
		{"until inclusive", models.GetProjectsRequest{CreatedAtUntil: &stamp}, []string{"pr-a"}},
		{"injection", models.GetProjectsRequest{Search: "' OR 1=1 --"}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.SortBy = "name"
			list, e := r.FindAll(ctx, &tc.req, "pr-owner")
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.want, ids)
			n, e := r.Count(ctx, &tc.req, "pr-owner")
			require.NoError(t, e)
			require.Equal(t, len(tc.want), n)
		})
	}
	for _, sort := range []string{"name", "created_at", "name; DROP TABLE projects"} {
		list, e := r.FindAll(ctx, &models.GetProjectsRequest{PageRequest: pm.PageRequest{SortBy: sort, SortDir: "desc", PageSize: 1, PageNum: 2}}, "pr-owner")
		require.NoError(t, e)
		require.Len(t, list, 1)
		require.Equal(t, "pr-a", list[0].ID)
	}
	p, err := r.FindSystem(ctx)
	require.NoError(t, err)
	require.Equal(t, "pr-a", p.ID)
	p, err = r.FindByIDAndAccountID(ctx, "pr-a", "pr-other")
	require.NoError(t, err)
	require.Nil(t, p)
	p, err = r.FindByIDAndAccountID(ctx, "pr-a", "pr-owner")
	require.NoError(t, err)
	require.Equal(t, &logo, p.LogoURL)
	p.Name = "Renamed"
	p.Description = nil
	p.LogoURL = nil
	p.IsActive = false
	n, err := r.Update(ctx, p)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	p, err = r.FindByID(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, "Renamed", p.Name)
	require.Nil(t, p.Description)
	require.Nil(t, p.LogoURL)
	require.False(t, p.IsActive)
	n, err = r.Delete(ctx, p)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	missing, err = r.FindByID(ctx, p.ID)
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByIDAndAccountID(ctx, p.ID, "pr-owner")
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindSystem(ctx)
	require.NoError(t, err)
	require.Nil(t, missing)
	n, err = r.Update(ctx, p)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = r.Delete(ctx, p)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = r.Count(ctx, &models.GetProjectsRequest{}, "pr-owner")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	foreign, err := r.FindByID(ctx, "pr-c")
	require.NoError(t, err)
	require.NotNil(t, foreign)
	var deleted int
	require.NoError(t, db.GetContext(ctx, &deleted, "SELECT COUNT(*) FROM projects WHERE id=? AND deleted_at IS NOT NULL", p.ID))
	require.Equal(t, 1, deleted)
}

func TestProjectRepositoryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := testSuite.GetDB()
	p := NewProjectRepository(db)
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"project create", func() error { return p.Create(ctx, &entities.Project{}) }},
		{"project count", func() error { _, e := p.Count(ctx, &models.GetProjectsRequest{}, "a"); return e }},
		{"project list", func() error { _, e := p.FindAll(ctx, &models.GetProjectsRequest{}, "a"); return e }},
		{"project id", func() error { _, e := p.FindByID(ctx, "p"); return e }},
		{"project account", func() error { _, e := p.FindByIDAndAccountID(ctx, "p", "a"); return e }},
		{"project system", func() error { _, e := p.FindSystem(ctx); return e }},
		{"project update", func() error { _, e := p.Update(ctx, &entities.Project{}); return e }},
		{"project delete", func() error { _, e := p.Delete(ctx, &entities.Project{}); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) { require.ErrorIs(t, tc.run(), context.Canceled) })
	}
}
