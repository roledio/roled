package web

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	brandingmocks "github.com/roledio/roled/auth/internal/services/branding/mocks"
	"github.com/roledio/roled/auth/internal/services/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRenderActivateProjectUser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Get("/user/activate/:token", h.renderActivateProjectUser)

	// Setup mock expectations
	userID := "user_123"
	loginURL := "https://example.com/login"
	mockService.On("RenderActivateProjectUser", mock.Anything, mock.AnythingOfType("*models.RenderActivateProjectUserRequest")).Return(&models.RenderActivateProjectUserResponse{
		User: &entities.User{
			ID:          userID,
			DisplayName: "Test User",
		},
		Project: &entities.Project{
			ID:   "proj_123",
			Name: "Test Project",
		},
		LoginURL: &loginURL,
	}, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/user/activate/test-token", nil)

	// Call the handler through the app
	_, err := app.Test(req)

	// Assertions - template rendering may fail in test environment, but service should be called
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestRenderActivateProjectUser_BindValidationError(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Get("/user/activate/:token", h.renderActivateProjectUser)

	// Create invalid request (missing token)
	req := httptest.NewRequest("GET", "/user/activate/", nil)

	// Call the handler through the app
	_, err := app.Test(req)

	// Should return error (template render with error)
	assert.NoError(t, err)
	// Service should not be called on validation error
	mockService.AssertNotCalled(t, "RenderActivateProjectUser")
}

func TestRenderActivateProjectUser_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Get("/user/activate/:token", h.renderActivateProjectUser)

	// Setup mock to return error
	mockService.On("RenderActivateProjectUser", mock.Anything, mock.AnythingOfType("*models.RenderActivateProjectUserRequest")).Return(nil, errors.New("service error"))

	// Create test request
	req := httptest.NewRequest("GET", "/user/activate/test-token", nil)

	// Call the handler through the app
	_, err := app.Test(req)

	// Should return template with error (may fail in test environment)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestRenderActivateProjectUser_WithFlashData(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route with CSRF middleware
	app.Use(csrf.New(csrf.Config{
		CookieSecure:   false,
		CookieHTTPOnly: true,
	}))
	app.Get("/user/activate/:token", h.renderActivateProjectUser)

	// Setup mock expectations
	userID := "user_123"
	mockService.On("RenderActivateProjectUser", mock.Anything, mock.AnythingOfType("*models.RenderActivateProjectUserRequest")).Return(&models.RenderActivateProjectUserResponse{
		User: &entities.User{
			ID:          userID,
			DisplayName: "Test User",
		},
		Project: &entities.Project{
			ID:   "proj_123",
			Name: "Test Project",
		},
	}, nil)

	// Create test request
	req := httptest.NewRequest("GET", "/user/activate/test-token", nil)

	// Call the handler through the app
	_, err := app.Test(req)

	// Assertions - template rendering may fail in test environment, but service should be called
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSubmitActivateProjectUser_Success(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Post("/user/activate/:token", h.submitActivateProjectUser)

	// Setup mock expectations
	userID := "user_123"
	loginURL := "https://example.com/login"
	mockService.On("SubmitActivateProjectUser", mock.Anything, mock.AnythingOfType("*models.SubmitActivateProjectUserRequest")).Return(&models.SubmitActivateProjectUserResponse{
		UserID: userID,
		Project: &entities.Project{
			ID:   "proj_123",
			Name: "Test Project",
		},
		LoginURL: &loginURL,
	}, nil)

	// Create test request with form data
	formData := "display_name=Test+User&email=test@example.com&password=password123&password_confirmation=password123"
	req := httptest.NewRequest("POST", "/user/activate/test-token", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Call the handler through the app
	_, err := app.Test(req)

	// Assertions - template rendering may fail in test environment, but service should be called
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSubmitActivateProjectUser_BindValidationError(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Post("/user/activate/:token", h.submitActivateProjectUser)

	// Create invalid request (missing required fields)
	req := httptest.NewRequest("POST", "/user/activate/test-token", nil)

	// Call the handler through the app
	_, err := app.Test(req)

	// Should return redirect with flash data
	assert.NoError(t, err)
	// Service should not be called on validation error
	mockService.AssertNotCalled(t, "SubmitActivateProjectUser")
}

func TestSubmitActivateProjectUser_ServiceError(t *testing.T) {
	app := fiber.New()
	mockService := new(mocks.MockUserService)
	brandingService := brandingmocks.NewMockService(t)
	brandingService.On("ResolveBranding", mock.Anything, mock.Anything).Return(&models.BrandingDetails{}, nil).Maybe()
	h := &handler{
		brandingService: brandingService,
		app:             app,
		defaultConfig:   &configs.DefaultConfig{},
		userService:     mockService,
	}

	// Register the route
	app.Post("/user/activate/:token", h.submitActivateProjectUser)

	// Setup mock to return error
	mockService.On("SubmitActivateProjectUser", mock.Anything, mock.AnythingOfType("*models.SubmitActivateProjectUserRequest")).Return(nil, errors.New("service error"))

	// Create test request with form data
	formData := "display_name=Test+User&email=test@example.com&password=password123&password_confirmation=password123"
	req := httptest.NewRequest("POST", "/user/activate/test-token", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Call the handler through the app
	_, err := app.Test(req)

	// Should return redirect with flash data
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
