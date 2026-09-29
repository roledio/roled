package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	rolemocks "github.com/roledio/roled/auth/internal/services/role/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoleHandlers(t *testing.T) {
	description := "Role description"
	role := models.RoleDetails{ID: "role-id", Name: "Role", Code: "reader"}
	details := &models.RoleDetailsAndPermissions{
		RoleDetails: role,
		Permissions: []models.RolePermission{{ID: "permission-id", ResourceName: "Users", PermissionName: "Read"}},
	}
	for _, tc := range []struct {
		name, method, route, url, body, serviceMethod string
		handle                                        func(*handler) fiber.Handler
		request                                       any
		result                                        any
	}{
		{
			name: "list", method: http.MethodGet, route: "/projects/:project_id/roles",
			url:           "/projects/project-id/roles?page_num=2&page_size=5&search=Role&sort_by=name&sort_dir=desc",
			serviceMethod: "GetRoles", handle: func(h *handler) fiber.Handler { return h.getProjectRoles },
			request: &models.GetProjectRolesRequest{ProjectID: "project-id", Search: "Role", PageRequest: pkgmodels.PageRequest{PageNum: 2, PageSize: 5, SortBy: "name", SortDir: "desc"}},
			result:  []models.RoleDetails{role},
		},
		{
			name: "details", method: http.MethodGet, route: "/projects/:project_id/roles/:role_id", url: "/projects/project-id/roles/role-id",
			serviceMethod: "GetRoleDetails", handle: func(h *handler) fiber.Handler { return h.getProjectRoleDetails },
			request: &models.GetRoleDetailsRequest{ProjectID: "project-id", RoleID: "role-id"}, result: &role,
		},
		{
			name: "create", method: http.MethodPost, route: "/projects/:project_id/roles", url: "/projects/project-id/roles",
			body:          `{"name":"Role","code":"reader","description":"Role description","permission_ids":["permission-id"]}`,
			serviceMethod: "CreateRole", handle: func(h *handler) fiber.Handler { return h.createProjectRole },
			request: &models.CreateRoleRequest{ProjectID: "project-id", Name: "Role", Code: "reader", Description: description, PermissionIDs: []string{"permission-id"}}, result: details,
		},
		{
			name: "update permissions", method: http.MethodPut, route: "/projects/:project_id/roles/:role_id", url: "/projects/project-id/roles/role-id",
			body:          `{"name":"Role","code":"reader","permission_ids":["permission-id"]}`,
			serviceMethod: "UpdateRole", handle: func(h *handler) fiber.Handler { return h.updateProjectRole },
			request: &models.UpdateRoleRequest{ProjectID: "project-id", RoleID: "role-id", Name: "Role", Code: "reader", PermissionIDs: []string{"permission-id"}}, result: details,
		},
		{
			name: "delete", method: http.MethodDelete, route: "/projects/:project_id/roles/:role_id", url: "/projects/project-id/roles/role-id",
			serviceMethod: "DeleteRole", handle: func(h *handler) fiber.Handler { return h.deleteProjectRole },
			request: &models.DeleteRoleRequest{ProjectID: "project-id", RoleID: "role-id"},
		},
	} {
		for _, serviceFails := range []bool{false, true} {
			outcome := "success"
			if serviceFails {
				outcome = "service error"
			}
			t.Run(tc.name+"/"+outcome, func(t *testing.T) {
				service := rolemocks.NewMockRoleService(t)
				var serviceErr error
				if serviceFails {
					serviceErr = domainerrors.ErrRoleNotFound
				}
				call := service.On(tc.serviceMethod, mock.Anything, tc.request).Once()
				switch tc.serviceMethod {
				case "GetRoles":
					call.Return(tc.result, 11, serviceErr)
				case "DeleteRole":
					call.Return(serviceErr)
				default:
					call.Return(tc.result, serviceErr)
				}
				app := fiber.New()
				app.Add([]string{tc.method}, tc.route, tc.handle(&handler{roleService: service}))
				req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
				res, err := app.Test(req)
				require.NoError(t, err)
				defer func() { require.NoError(t, res.Body.Close()) }()
				var body struct {
					Success    bool                  `json:"success"`
					Data       json.RawMessage       `json:"data"`
					Error      *pkgmodels.ErrorBody  `json:"error"`
					Pagination *pkgmodels.Pagination `json:"pagination"`
				}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
				if serviceFails {
					assert.Equal(t, http.StatusNotFound, res.StatusCode)
					assert.False(t, body.Success)
					require.NotNil(t, body.Error)
					assert.Equal(t, domainerrors.ErrRoleNotFound.Code, body.Error.Code)
					assert.Equal(t, domainerrors.ErrRoleNotFound.Msg, body.Error.Message)
					assert.Empty(t, body.Data)
					return
				}
				assert.Equal(t, http.StatusOK, res.StatusCode)
				assert.True(t, body.Success)
				assert.Nil(t, body.Error)
				if tc.result == nil {
					assert.Empty(t, body.Data)
				} else {
					want, err := json.Marshal(tc.result)
					require.NoError(t, err)
					assert.JSONEq(t, string(want), string(body.Data))
				}
				if tc.serviceMethod == "GetRoles" {
					assert.Equal(t, &pkgmodels.Pagination{PageNum: 2, PageSize: 1, TotalData: 11}, body.Pagination)
				} else {
					assert.Nil(t, body.Pagination)
				}
			})
		}
	}
}

func TestRoleHandlersRejectInvalidRequests(t *testing.T) {
	for _, tc := range []struct {
		name, method, route, body string
		handle                    func(*handler) fiber.Handler
	}{
		{"list missing project", http.MethodGet, "/roles", "", func(h *handler) fiber.Handler { return h.getProjectRoles }},
		{"details missing IDs", http.MethodGet, "/roles", "", func(h *handler) fiber.Handler { return h.getProjectRoleDetails }},
		{"delete missing IDs", http.MethodDelete, "/roles", "", func(h *handler) fiber.Handler { return h.deleteProjectRole }},
		{"create blank name", http.MethodPost, "/projects/:project_id/roles", `{"name":"  ","code":"reader"}`, func(h *handler) fiber.Handler { return h.createProjectRole }},
		{"create long name", http.MethodPost, "/projects/:project_id/roles", `{"name":"` + strings.Repeat("a", 51) + `","code":"reader"}`, func(h *handler) fiber.Handler { return h.createProjectRole }},
		{"create invalid code", http.MethodPost, "/projects/:project_id/roles", `{"name":"Reader","code":"invalid code!"}`, func(h *handler) fiber.Handler { return h.createProjectRole }},
		{"create malformed body", http.MethodPost, "/projects/:project_id/roles", `{`, func(h *handler) fiber.Handler { return h.createProjectRole }},
		{"update missing code", http.MethodPut, "/projects/:project_id/roles/:role_id", `{"name":"Role"}`, func(h *handler) fiber.Handler { return h.updateProjectRole }},
		{"update long description", http.MethodPut, "/projects/:project_id/roles/:role_id", `{"name":"Role","code":"reader","description":"` + strings.Repeat("a", 201) + `"}`, func(h *handler) fiber.Handler { return h.updateProjectRole }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := rolemocks.NewMockRoleService(t)
			app := fiber.New()
			app.Add([]string{tc.method}, tc.route, tc.handle(&handler{roleService: service}))
			url := strings.NewReplacer(":project_id", "project-id", ":role_id", "role-id").Replace(tc.route)
			req := httptest.NewRequest(tc.method, url, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var body pkgmodels.ResponseBody
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.False(t, body.Success)
			require.NotNil(t, body.Error)
			assert.Equal(t, pkgerrors.ErrInvalidParams.Code, body.Error.Code)
			assert.Empty(t, service.Calls)
		})
	}
}
