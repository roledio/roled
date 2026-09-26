package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	customerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	usermocks "github.com/roledio/roled/auth/internal/services/user/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProjectUserDetails(t *testing.T) {
	for _, tc := range []struct {
		name         string
		serviceError error
	}{
		{name: "success"},
		{name: "user not found", serviceError: customerrors.ErrUserNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := usermocks.NewMockUserService(t)
			user := &models.UserDetails{ID: "user-id", DisplayName: "Test User", IsActive: true, Permissions: []string{"users:read"}}
			service.EXPECT().GetUserDetails(mock.Anything, &models.GetUserDetailsRequest{
				ProjectID: "project-id", UserID: "user-id", IncludePermissions: true,
			}).Return(user, tc.serviceError).Once()
			h := &handler{userService: service}
			app := fiber.New()
			app.Get("/projects/:project_id/users/:user_id", h.getProjectUserDetails)
			res, err := app.Test(httptest.NewRequest(http.MethodGet, "/projects/project-id/users/user-id?include_permissions=true", nil))
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var body struct {
				Success bool               `json:"success"`
				Data    models.UserDetails `json:"data"`
				Error   struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
			if tc.serviceError != nil {
				assert.Equal(t, http.StatusNotFound, res.StatusCode)
				assert.False(t, body.Success)
				assert.Equal(t, customerrors.ErrUserNotFound.Code, body.Error.Code)
				assert.Equal(t, customerrors.ErrUserNotFound.Msg, body.Error.Message)
				return
			}
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.True(t, body.Success)
			assert.Equal(t, *user, body.Data)
		})
	}
}

func TestGetCurrentUserDetails(t *testing.T) {
	for _, tc := range []struct {
		name               string
		query              string
		includePermissions bool
		serviceError       error
	}{
		{name: "defaults to no permissions"},
		{name: "includes permissions", query: "?include_permissions=true", includePermissions: true},
		{name: "explicit false", query: "?include_permissions=false"},
		{name: "invalid boolean defaults to false", query: "?include_permissions=invalid"},
		{name: "service error", serviceError: customerrors.ErrUserNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := usermocks.NewMockUserService(t)
			user := &models.UserDetails{ID: "current-user", DisplayName: "Current User", IsActive: true}
			if tc.includePermissions {
				user.Permissions = []string{"users:read"}
			}
			service.EXPECT().GetCurrentUserDetails(mock.Anything, tc.includePermissions).Return(user, tc.serviceError).Once()
			h := &handler{userService: service}
			app := fiber.New()
			app.Get("/users/me", h.getCurrentUserDetails)
			res, err := app.Test(httptest.NewRequest(http.MethodGet, "/users/me"+tc.query, nil))
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var body struct {
				Success bool               `json:"success"`
				Data    models.UserDetails `json:"data"`
				Error   struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
			if tc.serviceError != nil {
				assert.Equal(t, http.StatusNotFound, res.StatusCode)
				assert.False(t, body.Success)
				assert.Equal(t, customerrors.ErrUserNotFound.Code, body.Error.Code)
				assert.Equal(t, customerrors.ErrUserNotFound.Msg, body.Error.Message)
				return
			}
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.True(t, body.Success)
			assert.Equal(t, *user, body.Data)
		})
	}
}

func TestCreateProjectUser(t *testing.T) {
	for _, tc := range []struct {
		name         string
		body         string
		request      *models.CreateUserRequest
		serviceError error
		wantError    *pkgerrors.CustomError
	}{
		{
			name:    "email and password",
			body:    `{"display_name":"Test User","email":"user@example.com","password":"password123","role_id":"role-id"}`,
			request: &models.CreateUserRequest{ProjectID: "project-id", DisplayName: "Test User", Email: "user@example.com", Password: "password123", RoleID: "role-id"},
		},
		{
			name:    "external user without password",
			body:    `{"display_name":"External User","external_user_id":"external-id"}`,
			request: &models.CreateUserRequest{ProjectID: "project-id", DisplayName: "External User", ExternalUserID: "external-id"},
		},
		{
			name:         "duplicate email",
			body:         `{"display_name":"Test User","email":"user@example.com","password":"password123"}`,
			request:      &models.CreateUserRequest{ProjectID: "project-id", DisplayName: "Test User", Email: "user@example.com", Password: "password123"},
			serviceError: customerrors.ErrUserEmailAlreadyUsed, wantError: &customerrors.ErrUserEmailAlreadyUsed,
		},
		{name: "missing display name", body: `{"external_user_id":"external-id"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "blank display name", body: `{"display_name":"   ","external_user_id":"external-id"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "missing identity", body: `{"display_name":"Test User"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "invalid email", body: `{"display_name":"Test User","email":"invalid","password":"password123"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "email requires password", body: `{"display_name":"Test User","email":"user@example.com","external_user_id":"external-id"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "short password", body: `{"display_name":"Test User","email":"user@example.com","password":"short"}`, wantError: &pkgerrors.ErrInvalidParams},
		{name: "invalid avatar URL", body: `{"display_name":"Test User","external_user_id":"external-id","avatar_url":"invalid"}`, wantError: &pkgerrors.ErrInvalidParams},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := usermocks.NewMockUserService(t)
			user := &models.UserDetails{ID: "new-user", DisplayName: "Created User", IsActive: true}
			if tc.request != nil {
				service.EXPECT().CreateUser(mock.Anything, tc.request).Return(user, tc.serviceError).Once()
			}
			h := &handler{userService: service}
			app := fiber.New()
			app.Post("/projects/:project_id/users", h.createProjectUser)
			req := httptest.NewRequest(http.MethodPost, "/projects/project-id/users", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			var body struct {
				Success bool               `json:"success"`
				Data    models.UserDetails `json:"data"`
				Error   struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
			if tc.wantError != nil {
				assert.Equal(t, tc.wantError.HttpCode, res.StatusCode)
				assert.False(t, body.Success)
				assert.Equal(t, tc.wantError.Code, body.Error.Code)
				assert.Equal(t, tc.wantError.Msg, body.Error.Message)
				if tc.request == nil {
					service.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
				}
				return
			}
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.True(t, body.Success)
			assert.Equal(t, *user, body.Data)
		})
	}
}
