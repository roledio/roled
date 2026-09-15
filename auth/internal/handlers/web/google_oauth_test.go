package web

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthorizeService is a mock implementation of AuthorizeService for testing
type MockAuthorizeService struct {
	mock.Mock
}

func (m *MockAuthorizeService) RenderAuthorize(ctx context.Context, req *models.RenderAuthorizeRequest) (*models.RenderAuthorizeResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RenderAuthorizeResult), args.Error(1)
}

func (m *MockAuthorizeService) SubmitAuthorize(ctx context.Context, req *models.SubmitAuthorizeRequest) (*models.SubmitAuthorizeResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubmitAuthorizeResult), args.Error(1)
}

func (m *MockAuthorizeService) InitiateGoogleOAuth(ctx context.Context, req *models.GoogleOAuthRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockAuthorizeService) HandleGoogleOAuthCallback(ctx context.Context, req *models.GoogleOAuthCallbackRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func TestHandleGoogleOAuth_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google", h.handleGoogleOAuth)

	// Setup mock expectations
	googleAuthURL := "https://accounts.google.com/o/oauth2/v2/auth?redirect_uri=..."
	mockService.On("InitiateGoogleOAuth", mock.Anything, mock.AnythingOfType("*models.GoogleOAuthRequest")).Return(googleAuthURL, nil)

	// Create test request with query parameters
	req := httptest.NewRequest("GET", "/oauth/google?client_id=test-client-id&redirect_uri=https://example.com/callback&response_type=code&code_challenge=test-challenge&code_challenge_method=S256&state=test-state&is_signup=true", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode) // Redirect status
	mockService.AssertExpectations(t)
}

func TestHandleGoogleOAuth_BindValidationError(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google", h.handleGoogleOAuth)

	// Create invalid request (missing required fields)
	req := httptest.NewRequest("GET", "/oauth/google?client_id=", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Should return redirect (with flash data)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	// Service should not be called on validation error
	mockService.AssertNotCalled(t, "InitiateGoogleOAuth")
}

func TestHandleGoogleOAuth_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google", h.handleGoogleOAuth)

	// Setup mock to return error
	mockService.On("InitiateGoogleOAuth", mock.Anything, mock.AnythingOfType("*models.GoogleOAuthRequest")).Return("", errors.New("service error"))

	// Create test request
	req := httptest.NewRequest("GET", "/oauth/google?client_id=test-client-id&redirect_uri=https://example.com/callback&response_type=code&code_challenge=test-challenge&code_challenge_method=S256&state=test-state", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Should return redirect (with flash data)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	mockService.AssertExpectations(t)
}

func TestHandleGoogleOAuthCallback_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google/callback", h.handleGoogleOAuthCallback)

	// Setup mock expectations
	redirectURL := "https://example.com/callback?code=auth_code&state=test_state"
	mockService.On("HandleGoogleOAuthCallback", mock.Anything, mock.AnythingOfType("*models.GoogleOAuthCallbackRequest")).Return(redirectURL, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/oauth/google/callback?code=test-auth-code&state=test-state", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode) // Redirect status
	mockService.AssertExpectations(t)
}

func TestHandleGoogleOAuthCallback_BindValidationError(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google/callback", h.handleGoogleOAuthCallback)

	// Create invalid request (missing required fields)
	req := httptest.NewRequest("GET", "/oauth/google/callback?code=", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Should return error
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	// Service should not be called on validation error
	mockService.AssertNotCalled(t, "HandleGoogleOAuthCallback")
}

func TestHandleGoogleOAuthCallback_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAuthorizeService)
	h := &handler{
		app:              app,
		defaultConfig:    &configs.DefaultConfig{},
		authorizeService: mockService,
	}

	// Register the route
	app.Get("/oauth/google/callback", h.handleGoogleOAuthCallback)

	// Setup mock to return error
	mockService.On("HandleGoogleOAuthCallback", mock.Anything, mock.AnythingOfType("*models.GoogleOAuthCallbackRequest")).Return("", errors.New("callback service error"))

	// Create test request
	req := httptest.NewRequest("GET", "/oauth/google/callback?code=test-auth-code&state=test-state", nil)

	// Call the handler through the app
	resp, err := app.Test(req)

	// Should return redirect to authorize page with error
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusSeeOther, resp.StatusCode)
	mockService.AssertExpectations(t)
}
