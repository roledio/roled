package mariadb

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	pm "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/require"
)

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

func TestRoleRepositoryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := testSuite.GetDB()
	r := NewRoleRepository(db).(*roleRepository)
	for _, tc := range []struct {
		name string
		run  func() error
	}{
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
