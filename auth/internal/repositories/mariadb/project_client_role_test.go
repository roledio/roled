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

func TestRoleRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	r := NewRoleRepository(db).(*roleRepository)
	testSuite.CleanTables(t, "user_roles", "users", "roles", "projects", "accounts")
	_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: "rr-owner", Name: "Owner", IsActive: true})
	require.NoError(t, err)
	for _, id := range []string{"rr-project", "rr-other"} {
		_, err = testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: id, AccountID: "rr-owner", Name: id, IsActive: true})
		require.NoError(t, err)
	}
	for _, v := range []entities.Role{
		{ID: "rr-a", AccountID: "rr-owner", ProjectID: "rr-project", Name: "Alpha", Code: "admin", Description: "special"},
		{ID: "rr-b", AccountID: "rr-owner", ProjectID: "rr-project", Name: "Beta", Code: "viewer"},
		{ID: "rr-c", AccountID: "rr-owner", ProjectID: "rr-other", Name: "Foreign", Code: "admin"},
	} {
		require.NoError(t, r.Create(ctx, &v))
	}
	for _, tc := range []struct {
		search string
		want   []string
	}{{"", []string{"rr-a", "rr-b"}}, {" Alpha ", []string{"rr-a"}}, {"admin", []string{"rr-a"}}, {"special", []string{"rr-a"}}, {"' OR 1=1 --", []string{}}} {
		t.Run("search/"+tc.search, func(t *testing.T) {
			req := &models.GetProjectRolesRequest{ProjectID: "rr-project", Search: tc.search, PageRequest: pm.PageRequest{SortBy: "name"}}
			list, e := r.FindAll(ctx, req)
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.want, ids)
			n, e := r.Count(ctx, req)
			require.NoError(t, e)
			require.Equal(t, len(tc.want), n)
		})
	}
	for _, sort := range []string{"name", "code", "created_at", "updated_at", "name; DROP TABLE roles"} {
		list, e := r.FindAll(ctx, &models.GetProjectRolesRequest{ProjectID: "rr-project", PageRequest: pm.PageRequest{SortBy: sort, SortDir: "desc", PageSize: 1, PageNum: 2}})
		require.NoError(t, e)
		require.Len(t, list, 1)
		if sort == "name" || sort == "code" {
			require.Equal(t, "rr-a", list[0].ID)
		}
	}
	role, err := r.FindByProjectIDAndCode(ctx, "rr-project", "admin")
	require.NoError(t, err)
	require.Equal(t, "rr-a", role.ID)
	missing, err := r.FindByIDAndProjectID(ctx, role.ID, "rr-other")
	require.NoError(t, err)
	require.Nil(t, missing)
	role, err = r.FindByIDAndProjectID(ctx, role.ID, "rr-project")
	require.NoError(t, err)
	role.Name = "Renamed"
	role.Code = "owner"
	role.Description = "changed"
	n, err := r.Update(ctx, role)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	role, err = r.FindByID(ctx, role.ID)
	require.NoError(t, err)
	require.Equal(t, "Renamed", role.Name)
	require.Equal(t, "owner", role.Code)
	require.Equal(t, "changed", role.Description)
	email := "member@example.com"
	_, err = testutil.CreateUser(ctx, db, testutil.UserFixture{ID: "rr-user", AccountID: "rr-owner", ProjectID: "rr-project", DisplayName: "User", Email: &email, IsActive: true})
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO user_roles(user_id,role_id) VALUES (?,?)", "rr-user", role.ID)
	require.NoError(t, err)
	joined, err := r.FindByUserID(ctx, "rr-user")
	require.NoError(t, err)
	require.Equal(t, role.ID, joined.ID)
	n, err = r.DeleteByID(ctx, role.ID)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	missing, err = r.FindByID(ctx, role.ID)
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByIDAndProjectID(ctx, role.ID, "rr-project")
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByProjectIDAndCode(ctx, "rr-project", "owner")
	require.NoError(t, err)
	require.Nil(t, missing)
	missing, err = r.FindByUserID(ctx, "rr-user")
	require.NoError(t, err)
	require.Nil(t, missing)
	n, err = r.Update(ctx, role)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = r.Count(ctx, &models.GetProjectRolesRequest{ProjectID: "rr-project"})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	foreign, err := r.FindByID(ctx, "rr-c")
	require.NoError(t, err)
	require.NotNil(t, foreign)
	var deleted int
	require.NoError(t, db.GetContext(ctx, &deleted, "SELECT COUNT(*) FROM roles WHERE id=? AND deleted_at IS NOT NULL", role.ID))
	require.Equal(t, 1, deleted)
}

func TestProjectClientRoleCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := testSuite.GetDB()
	p := NewProjectRepository(db)
	c := NewClientRepository(db)
	r := NewRoleRepository(db).(*roleRepository)
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
		{"client create", func() error { return c.Create(ctx, &entities.Client{}) }},
		{"client count", func() error { _, e := c.Count(ctx, &models.GetClientsRequest{}); return e }},
		{"client list", func() error { _, e := c.FindAll(ctx, &models.GetClientsRequest{}); return e }},
		{"client id", func() error { _, e := c.FindByID(ctx, "c"); return e }},
		{"client project", func() error { _, e := c.FindByIDAndProjectID(ctx, "c", "p"); return e }},
		{"client default", func() error { _, e := c.FindByProjectIDAndIsDefault(ctx, "p", true); return e }},
		{"client update", func() error { _, e := c.Update(ctx, &entities.Client{}); return e }},
		{"client delete", func() error { _, e := c.Delete(ctx, &entities.Client{}); return e }},
		{"client delete project", func() error { _, e := c.DeleteByProjectID(ctx, "p"); return e }},
		{"role create", func() error { return r.Create(ctx, &entities.Role{}) }},
		{"role count", func() error { _, e := r.Count(ctx, &models.GetProjectRolesRequest{}); return e }},
		{"role list", func() error { _, e := r.FindAll(ctx, &models.GetProjectRolesRequest{}); return e }},
		{"role id", func() error { _, e := r.FindByID(ctx, "r"); return e }},
		{"role project", func() error { _, e := r.FindByIDAndProjectID(ctx, "r", "p"); return e }},
		{"role code", func() error { _, e := r.FindByProjectIDAndCode(ctx, "p", "r"); return e }},
		{"role user", func() error { _, e := r.FindByUserID(ctx, "u"); return e }},
		{"role update", func() error { _, e := r.Update(ctx, &entities.Role{}); return e }},
		{"role delete", func() error { _, e := r.DeleteByID(ctx, "r"); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) { require.ErrorIs(t, tc.run(), context.Canceled) })
	}
}
