package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tidwall/gjson"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/pkg/errors"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
)

func testHandler(c fiber.Ctx) error {
	c.Status(http.StatusOK)
	return nil
}

func TestPermission_NoAccessToken(t *testing.T) {
	app := fiber.New()

	mockRegistry := repositorymocks.NewMockRegistry(t)

	// Add permission middleware to this route with name: test
	app.Get("/test", Permission(mockRegistry), testHandler).Name("test")

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req)

	expectedErr := errors.ErrInvalidAuthorizationToken

	assert.NoError(t, err)
	assert.Equal(t, expectedErr.HttpCode, res.StatusCode)

	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, expectedErr.Code, parsed.Get("error.code").String())
	assert.Equal(t, expectedErr.Msg, parsed.Get("error.message").String())
}

func TestPermission_RouteNoRequiredPermissions(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)

	// Mock registry - but it shouldn't be called for routes without required permissions
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Maybe()
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo).Maybe()

	// Mock access token with user ID
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   pointy.String("user-id"),
	}

	// Mock permissions - but shouldn't be called
	mockRoleRepo.EXPECT().FindByUserID(mock.Anything, "user-id").Return(&entities.Role{ID: "role-id"}, nil).Maybe()
	mockPermissionRepo.EXPECT().FindByRoleID(mock.Anything, "role-id").Return([]interfaces.PermissionResource{}, nil).Maybe()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name("test")

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPermission_UserHasPermission(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo).Once()

	// Mock access token with user ID
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   pointy.String("user-id"),
	}

	// Mock role
	mockRoleRepo.EXPECT().FindByUserID(mock.Anything, "user-id").Return(&entities.Role{ID: "role-id"}, nil).Once()

	// Mock permissions
	permissions := []interfaces.PermissionResource{
		{
			ResourceCode: "accounts",
			Code:         "read",
		},
	}
	mockPermissionRepo.EXPECT().FindByRoleID(mock.Anything, "role-id").Return(permissions, nil).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPermission_UserNoPermission(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo).Once()

	// Mock access token with user ID
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   pointy.String("user-id"),
	}

	// Mock role
	mockRoleRepo.EXPECT().FindByUserID(mock.Anything, "user-id").Return(&entities.Role{ID: "role-id"}, nil).Once()

	// Mock empty permissions
	permissions := []interfaces.PermissionResource{}
	mockPermissionRepo.EXPECT().FindByRoleID(mock.Anything, "role-id").Return(permissions, nil).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req)

	expectedErr := errors.ErrInsufficientPermission

	assert.NoError(t, err)
	assert.Equal(t, expectedErr.HttpCode, res.StatusCode)

	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, expectedErr.Code, parsed.Get("error.code").String())
	assert.Equal(t, expectedErr.Msg, parsed.Get("error.message").String())
}

func TestPermission_ClientHasPermission(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()

	// Mock access token without user ID (client token)
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   nil,
	}

	// Mock permissions
	permissions := []interfaces.PermissionResource{
		{
			ResourceCode: "accounts",
			Code:         "read",
		},
	}
	mockPermissionRepo.EXPECT().FindByClientID(mock.Anything, "client-id").Return(permissions, nil).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPermission_ClientNoPermission(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()

	// Mock access token without user ID (client token)
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   nil,
	}

	// Mock empty permissions
	permissions := []interfaces.PermissionResource{}
	mockPermissionRepo.EXPECT().FindByClientID(mock.Anything, "client-id").Return(permissions, nil).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req)

	expectedErr := errors.ErrInsufficientPermission

	assert.NoError(t, err)
	assert.Equal(t, expectedErr.HttpCode, res.StatusCode)

	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, expectedErr.Code, parsed.Get("error.code").String())
	assert.Equal(t, expectedErr.Msg, parsed.Get("error.message").String())
}

func TestPermission_FindByUserIDError(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo).Once()

	// Mock access token with user ID
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   pointy.String("user-id"),
	}

	// Mock repository error
	mockRoleRepo.EXPECT().FindByUserID(mock.Anything, "user-id").Return(nil, assert.AnError).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req)

	expectedErr := errors.ErrSystemError

	assert.NoError(t, err)
	assert.Equal(t, expectedErr.HttpCode, res.StatusCode)

	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, expectedErr.Code, parsed.Get("error.code").String())
	assert.Equal(t, expectedErr.Msg, parsed.Get("error.message").String())
}

func TestPermission_FindByClientIDError(t *testing.T) {
	app := fiber.New()
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockPermissionRepo := repositorymocks.NewMockPermissionRepository(t)

	// Mock registry
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo).Once()

	// Mock access token with user ID
	accessToken := &entities.AccessToken{
		ID:       "token-id",
		ClientID: "client-id",
		UserID:   nil,
	}

	// Mock repository error
	mockPermissionRepo.EXPECT().FindByClientID(mock.Anything, "client-id").Return(nil, assert.AnError).Once()

	// Setup middleware chain: first set token, then check permission
	app.Use(func(c fiber.Ctx) error {
		c.Locals(constants.CtxAccessToken, accessToken)
		return c.Next()
	})

	// Add route handler
	app.Get("/test", Permission(mockRegistry), testHandler).Name(constants.RouteGetAccounts)

	req := httptest.NewRequest("GET", "/test", nil)
	res, err := app.Test(req)

	expectedErr := errors.ErrSystemError

	assert.NoError(t, err)
	assert.Equal(t, expectedErr.HttpCode, res.StatusCode)

	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, expectedErr.Code, parsed.Get("error.code").String())
	assert.Equal(t, expectedErr.Msg, parsed.Get("error.message").String())
}
