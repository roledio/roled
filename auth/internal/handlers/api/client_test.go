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
	clientmocks "github.com/roledio/roled/auth/internal/services/client/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientHandlers(t *testing.T) {
	active := false
	description := "Client description"
	client := models.ClientDetails{ID: "client-id", Name: "Client", IsActive: true}
	details := &models.ClientDetailsAndPermissions{
		ClientDetails: client,
		Permissions:   []models.ClientPermission{{ID: "permission-id", ResourceName: "Users", PermissionName: "Read"}},
	}
	for _, tc := range []struct {
		name, method, route, url, body, serviceMethod string
		handle                                        func(*handler) fiber.Handler
		request                                       any
		result                                        any
	}{
		{
			name: "list", method: http.MethodGet, route: "/projects/:project_id/clients",
			url:           "/projects/project-id/clients?page_num=2&page_size=5&search=Client&is_active=false&sort_by=name&sort_dir=desc",
			serviceMethod: "GetClients", handle: func(h *handler) fiber.Handler { return h.getProjectClients },
			request: &models.GetClientsRequest{ProjectID: "project-id", Search: "Client", IsActive: &active, PageRequest: pkgmodels.PageRequest{PageNum: 2, PageSize: 5, SortBy: "name", SortDir: "desc"}},
			result:  []models.ClientDetails{client},
		},
		{
			name: "details", method: http.MethodGet, route: "/projects/:project_id/clients/:client_id", url: "/projects/project-id/clients/client-id",
			serviceMethod: "GetClientDetails", handle: func(h *handler) fiber.Handler { return h.getProjectClientDetails },
			request: &models.GetClientDetailsRequest{ProjectID: "project-id", ClientID: "client-id"}, result: &client,
		},
		{
			name: "create", method: http.MethodPost, route: "/projects/:project_id/clients", url: "/projects/project-id/clients",
			body:          `{"name":"Client","description":"Client description","permission_ids":["permission-id"]}`,
			serviceMethod: "CreateClient", handle: func(h *handler) fiber.Handler { return h.createProjectClient },
			request: &models.CreateClientRequest{ProjectID: "project-id", Name: "Client", Description: &description, PermissionIDs: []string{"permission-id"}}, result: details,
		},
		{
			name: "update inactive", method: http.MethodPut, route: "/projects/:project_id/clients/:client_id", url: "/projects/project-id/clients/client-id",
			body:          `{"name":"Client","is_active":false,"permission_ids":["permission-id"]}`,
			serviceMethod: "UpdateClient", handle: func(h *handler) fiber.Handler { return h.updateProjectClient },
			request: &models.UpdateClientRequest{ProjectID: "project-id", ClientID: "client-id", Name: "Client", IsActive: &active, PermissionIDs: []string{"permission-id"}}, result: details,
		},
		{
			name: "delete", method: http.MethodDelete, route: "/projects/:project_id/clients/:client_id", url: "/projects/project-id/clients/client-id",
			serviceMethod: "DeleteClient", handle: func(h *handler) fiber.Handler { return h.deleteProjectClient },
			request: &models.DeleteClientRequest{ProjectID: "project-id", ClientID: "client-id"},
		},
	} {
		for _, serviceFails := range []bool{false, true} {
			outcome := "success"
			if serviceFails {
				outcome = "service error"
			}
			t.Run(tc.name+"/"+outcome, func(t *testing.T) {
				service := clientmocks.NewMockClientService(t)
				var serviceErr error
				if serviceFails {
					serviceErr = domainerrors.ErrClientNotFound
				}
				call := service.On(tc.serviceMethod, mock.Anything, tc.request).Once()
				switch tc.serviceMethod {
				case "GetClients":
					call.Return(tc.result, 11, serviceErr)
				case "DeleteClient":
					call.Return(serviceErr)
				default:
					call.Return(tc.result, serviceErr)
				}
				app := fiber.New()
				app.Add([]string{tc.method}, tc.route, tc.handle(&handler{clientService: service}))
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
					assert.Equal(t, domainerrors.ErrClientNotFound.Code, body.Error.Code)
					assert.Equal(t, domainerrors.ErrClientNotFound.Msg, body.Error.Message)
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
				if tc.serviceMethod == "GetClients" {
					assert.Equal(t, &pkgmodels.Pagination{PageNum: 2, PageSize: 1, TotalData: 11}, body.Pagination)
				} else {
					assert.Nil(t, body.Pagination)
				}
			})
		}
	}
}

func TestClientHandlersRejectInvalidRequests(t *testing.T) {
	for _, tc := range []struct {
		name, method, route, body string
		handle                    func(*handler) fiber.Handler
	}{
		{"list missing project", http.MethodGet, "/clients", "", func(h *handler) fiber.Handler { return h.getProjectClients }},
		{"details missing IDs", http.MethodGet, "/clients", "", func(h *handler) fiber.Handler { return h.getProjectClientDetails }},
		{"delete missing IDs", http.MethodDelete, "/clients", "", func(h *handler) fiber.Handler { return h.deleteProjectClient }},
		{"create blank name", http.MethodPost, "/projects/:project_id/clients", `{"name":"  "}`, func(h *handler) fiber.Handler { return h.createProjectClient }},
		{"create long name", http.MethodPost, "/projects/:project_id/clients", `{"name":"` + strings.Repeat("a", 51) + `"}`, func(h *handler) fiber.Handler { return h.createProjectClient }},
		{"create malformed body", http.MethodPost, "/projects/:project_id/clients", `{`, func(h *handler) fiber.Handler { return h.createProjectClient }},
		{"update missing active", http.MethodPut, "/projects/:project_id/clients/:client_id", `{"name":"Client"}`, func(h *handler) fiber.Handler { return h.updateProjectClient }},
		{"update long description", http.MethodPut, "/projects/:project_id/clients/:client_id", `{"name":"Client","is_active":true,"description":"` + strings.Repeat("a", 201) + `"}`, func(h *handler) fiber.Handler { return h.updateProjectClient }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := clientmocks.NewMockClientService(t)
			app := fiber.New()
			app.Add([]string{tc.method}, tc.route, tc.handle(&handler{clientService: service}))
			url := strings.NewReplacer(":project_id", "project-id", ":client_id", "client-id").Replace(tc.route)
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
