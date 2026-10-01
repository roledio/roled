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

func TestResourcePermissionQueriesAndIsolation(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	testSuite.CleanTables(t, "client_permissions", "role_permissions", "permissions", "resources", "clients", "roles", "projects", "accounts")
	_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: "rp-owner", Name: "Owner", IsActive: true})
	require.NoError(t, err)
	for _, id := range []string{"rp-project", "rp-other"} {
		_, err = testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: id, AccountID: "rp-owner", Name: id, IsActive: true})
		require.NoError(t, err)
	}
	resources := NewResourceRepository(db).(*resourceRepository)
	permissions := NewPermissionRepository(db)
	desc := "invoice records"
	count, err := resources.Create(ctx, []entities.Resource{
		{ID: "rp-a", AccountID: "rp-owner", ProjectID: "rp-project", Name: "Alpha", Code: "alpha", Description: &desc},
		{ID: "rp-b", AccountID: "rp-owner", ProjectID: "rp-project", Name: "Beta", Code: "beta", IsDefault: true},
		{ID: "rp-c", AccountID: "rp-owner", ProjectID: "rp-other", Name: "Other", Code: "alpha"},
	})
	require.NoError(t, err)
	require.Equal(t, 3, count)
	note := "download report"
	count, err = permissions.Create(ctx, []entities.Permission{
		{ID: "rp-read", ResourceID: "rp-a", Name: "Read", Code: "read", Description: &note},
		{ID: "rp-write", ResourceID: "rp-b", Name: "Write", Code: "write", IsDefault: true},
		{ID: "rp-foreign", ResourceID: "rp-c", Name: "Foreign", Code: "read"},
	})
	require.NoError(t, err)
	require.Equal(t, 3, count)
	yes, no := true, false
	for _, tc := range []struct {
		name, search string
		filter       *bool
		expected     []string
	}{
		{"all", "", nil, []string{"rp-a", "rp-b"}},
		{"description", " invoice ", nil, []string{"rp-a"}},
		{"permission name", "Read", nil, []string{"rp-a"}},
		{"permission description", "download", nil, []string{"rp-a"}},
		{"default", "", &yes, []string{"rp-b"}},
		{"nondefault", "", &no, []string{"rp-a"}},
		{"injection search", "' OR 1=1 --", nil, []string{}},
	} {
		t.Run("resources/"+tc.name, func(t *testing.T) {
			req := &models.GetResourcesRequest{ProjectID: "rp-project", Search: tc.search, IsDefault: tc.filter}
			list, e := resources.FindAll(ctx, req)
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.expected, ids)
			n, e := resources.Count(ctx, req)
			require.NoError(t, e)
			require.Equal(t, len(tc.expected), n)
		})
	}
	for _, tc := range []struct {
		name, search string
		filter       *bool
		expected     []string
	}{
		{"all", "", nil, []string{"rp-read", "rp-write"}},
		{"resource name", " Alpha ", nil, []string{"rp-read"}},
		{"permission name", "Write", nil, []string{"rp-write"}},
		{"description", "download", nil, []string{"rp-read"}},
		{"default", "", &yes, []string{"rp-write"}},
		{"nondefault", "", &no, []string{"rp-read"}},
		{"injection search", "' OR 1=1 --", nil, []string{}},
	} {
		t.Run("permissions/"+tc.name, func(t *testing.T) {
			req := &models.GetPermissionsRequest{ProjectID: "rp-project", Search: tc.search, IsDefault: tc.filter}
			list, e := permissions.FindAll(ctx, req)
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.expected, ids)
			n, e := permissions.Count(ctx, req)
			require.NoError(t, e)
			require.Equal(t, len(tc.expected), n)
		})
	}
	for _, sort := range []string{"name", "code", "resource_name", "resource_code", "created_at", "updated_at", "name; DROP TABLE permissions"} {
		list, e := permissions.FindAll(ctx, &models.GetPermissionsRequest{ProjectID: "rp-project", PageRequest: pm.PageRequest{SortBy: sort, SortDir: "desc", PageSize: 1, PageNum: 2}})
		require.NoError(t, e)
		require.Len(t, list, 1)
		if sort == "name" || sort == "code" || sort == "resource_name" || sort == "resource_code" {
			require.Equal(t, "rp-read", list[0].ID)
		}
	}
	list, e := resources.FindAll(ctx, &models.GetResourcesRequest{ProjectID: "rp-project", PageRequest: pm.PageRequest{SortBy: "name", SortDir: "desc", PageSize: 1, PageNum: 2}})
	require.NoError(t, e)
	require.Len(t, list, 1)
	require.Equal(t, "rp-b", list[0].ID)
	projectResources, e := resources.FindByProjectID(ctx, "rp-project")
	require.NoError(t, e)
	require.Len(t, projectResources, 2)
	for _, filter := range []*bool{nil, &yes, &no} {
		found, e := permissions.FindByProjectID(ctx, "rp-project", filter)
		require.NoError(t, e)
		if filter == nil {
			require.Len(t, found, 2)
		} else {
			require.Len(t, found, 1)
			require.Equal(t, *filter, found[0].IsDefault)
		}
	}
	found, e := permissions.FindByResourceIDsAndSearch(ctx, []string{"rp-a", "rp-b"}, " download ")
	require.NoError(t, e)
	require.Len(t, found, 1)
	require.Equal(t, "rp-read", found[0].ID)
	found, e = permissions.FindByResourceIDsAndSearch(ctx, []string{"rp-a", "rp-b"}, "")
	require.NoError(t, e)
	require.Len(t, found, 2)
	joined, e := permissions.FindByIDs(ctx, []string{"rp-read"})
	require.NoError(t, e)
	require.Len(t, joined, 1)
	require.Equal(t, "Alpha", joined[0].ResourceName)
	require.Equal(t, "alpha", joined[0].ResourceCode)
	_, err = testutil.CreateRole(ctx, db, testutil.RoleFixture{ID: "rp-role", AccountID: "rp-owner", ProjectID: "rp-project", Name: "Role", Code: "role"})
	require.NoError(t, err)
	_, err = testutil.CreateClient(ctx, db, testutil.ClientFixture{ID: "rp-client", AccountID: "rp-owner", ProjectID: "rp-project", Name: "Client", IsActive: true})
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO role_permissions (role_id,permission_id) VALUES (?,?)", "rp-role", "rp-read")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO client_permissions (client_id,permission_id) VALUES (?,?)", "rp-client", "rp-write")
	require.NoError(t, err)
	joined, e = permissions.FindByRoleID(ctx, "rp-role")
	require.NoError(t, e)
	require.Len(t, joined, 1)
	require.Equal(t, "rp-read", joined[0].ID)
	joined, e = permissions.FindByClientID(ctx, "rp-client")
	require.NoError(t, e)
	require.Len(t, joined, 1)
	require.Equal(t, "rp-write", joined[0].ID)
	joined, e = permissions.FindByRoleID(ctx, "missing")
	require.NoError(t, e)
	require.Empty(t, joined)
	joined, e = permissions.FindByClientID(ctx, "missing")
	require.NoError(t, e)
	require.Empty(t, joined)
	resource, e := resources.FindByIDAndProjectID(ctx, "rp-a", "rp-other")
	require.NoError(t, e)
	require.Nil(t, resource)
	resource, e = resources.FindByProjectIDAndCode(ctx, "rp-project", "missing")
	require.NoError(t, e)
	require.Nil(t, resource)
	resource, e = resources.FindByProjectIDAndCode(ctx, "rp-project", "alpha")
	require.NoError(t, e)
	require.Equal(t, "rp-a", resource.ID)
	resource.Name = "Renamed"
	resource.Description = nil
	count, e = resources.Update(ctx, resource)
	require.NoError(t, e)
	require.Equal(t, 1, count)
	resource, e = resources.FindByIDAndProjectID(ctx, "rp-a", "rp-project")
	require.NoError(t, e)
	require.Equal(t, "Renamed", resource.Name)
	require.Nil(t, resource.Description)
	count, e = permissions.DeleteByResourceID(ctx, "rp-c")
	require.NoError(t, e)
	require.Equal(t, 1, count)
	foreign, e := resources.FindByIDAndProjectID(ctx, "rp-c", "rp-other")
	require.NoError(t, e)
	count, e = resources.Delete(ctx, foreign)
	require.NoError(t, e)
	require.Equal(t, 1, count)
	remaining, e := permissions.FindByProjectID(ctx, "rp-project", nil)
	require.NoError(t, e)
	require.Len(t, remaining, 2)
}

func TestResourcePermissionCancellationAndEmptyInputs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewResourceRepository(testSuite.GetDB()).(*resourceRepository)
	p := NewPermissionRepository(testSuite.GetDB())
	n, err := r.Create(ctx, nil)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = p.Create(ctx, nil)
	require.NoError(t, err)
	require.Zero(t, n)
	ps, err := p.FindByIDs(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, ps)
	empty, err := p.FindByResourceIDsAndSearch(ctx, nil, "")
	require.NoError(t, err)
	require.Empty(t, empty)
	checks := []struct {
		name string
		run  func() error
	}{
		{"resource create", func() error { _, e := r.Create(ctx, []entities.Resource{{ID: "cancel"}}); return e }},
		{"resource update", func() error { _, e := r.Update(ctx, &entities.Resource{ID: "cancel"}); return e }},
		{"resource delete", func() error { _, e := r.Delete(ctx, &entities.Resource{ID: "cancel"}); return e }},
		{"resource list", func() error { _, e := r.FindAll(ctx, &models.GetResourcesRequest{ProjectID: "p"}); return e }},
		{"resource count", func() error { _, e := r.Count(ctx, &models.GetResourcesRequest{ProjectID: "p"}); return e }},
		{"resource project", func() error { _, e := r.FindByProjectID(ctx, "p"); return e }},
		{"resource id", func() error { _, e := r.FindByIDAndProjectID(ctx, "r", "p"); return e }},
		{"resource code", func() error { _, e := r.FindByProjectIDAndCode(ctx, "p", "r"); return e }},
		{"permission create", func() error { _, e := p.Create(ctx, []entities.Permission{{ID: "cancel"}}); return e }},
		{"permission delete", func() error { _, e := p.DeleteByResourceID(ctx, "r"); return e }},
		{"permission list", func() error { _, e := p.FindAll(ctx, &models.GetPermissionsRequest{ProjectID: "p"}); return e }},
		{"permission count", func() error { _, e := p.Count(ctx, &models.GetPermissionsRequest{ProjectID: "p"}); return e }},
		{"permission project", func() error { _, e := p.FindByProjectID(ctx, "p", nil); return e }},
		{"permission resource", func() error { _, e := p.FindByResourceIDsAndSearch(ctx, []string{"r"}, ""); return e }},
		{"permission ids", func() error { _, e := p.FindByIDs(ctx, []string{"r"}); return e }},
		{"permission role", func() error { _, e := p.FindByRoleID(ctx, "r"); return e }},
		{"permission client", func() error { _, e := p.FindByClientID(ctx, "r"); return e }},
	}
	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) { require.ErrorIs(t, tc.run(), context.Canceled) })
	}
}
