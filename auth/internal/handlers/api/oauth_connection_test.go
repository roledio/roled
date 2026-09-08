package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	customerrors "github.com/roledio/roled/auth/internal/errors"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tidwall/gjson"
)

func TestGetProjectOAuthConnections_Success(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	now := time.Now()
	expectedResponse := []models.OAuthConnectionDetails{
		{
			ID:             "conn-1",
			ProjectID:      "proj-123",
			Provider:       "google",
			CredentialType: "default",
			Enabled:        true,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	oauthConnectionServiceMock.EXPECT().GetOAuthConnections(mock.Anything, mock.MatchedBy(func(req *models.GetOAuthConnectionsRequest) bool {
		return req.ProjectID == "proj-123"
	})).Return(expectedResponse, nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Get("/api/v1/projects/:project_id/oauth-connections", h.getOAuthConnections)

	req := httptest.NewRequest("GET", "/api/v1/projects/proj-123/oauth-connections", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "conn-1", parsed.Get("data.0.id").String())
	assert.Equal(t, "proj-123", parsed.Get("data.0.project_id").String())
	assert.Equal(t, "google", parsed.Get("data.0.provider").String())
	assert.Equal(t, "default", parsed.Get("data.0.credential_type").String())
	assert.True(t, parsed.Get("data.0.enabled").Bool())
}

func TestGetProjectOAuthConnections_EmptyList(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	expectedResponse := []models.OAuthConnectionDetails{}

	oauthConnectionServiceMock.EXPECT().GetOAuthConnections(mock.Anything, mock.MatchedBy(func(req *models.GetOAuthConnectionsRequest) bool {
		return req.ProjectID == "proj-123"
	})).Return(expectedResponse, nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Get("/api/v1/projects/:project_id/oauth-connections", h.getOAuthConnections)

	req := httptest.NewRequest("GET", "/api/v1/projects/proj-123/oauth-connections", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "[]", parsed.Get("data").Raw)
}

func TestGetProjectOAuthConnection_Success(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	now := time.Now()
	expectedResponse := &models.OAuthConnectionDetails{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: "custom",
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	oauthConnectionServiceMock.EXPECT().GetOAuthConnectionDetails(mock.Anything, mock.MatchedBy(func(req *models.GetOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" && req.Provider == "google"
	})).Return(expectedResponse, nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Get("/api/v1/projects/:project_id/oauth-connections/:provider", h.getOAuthConnectionDetails)

	req := httptest.NewRequest("GET", "/api/v1/projects/proj-123/oauth-connections/google", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "conn-1", parsed.Get("data.id").String())
	assert.Equal(t, "proj-123", parsed.Get("data.project_id").String())
	assert.Equal(t, "google", parsed.Get("data.provider").String())
	assert.Equal(t, "custom", parsed.Get("data.credential_type").String())
	assert.True(t, parsed.Get("data.enabled").Bool())
}

func TestGetProjectOAuthConnection_NotFound(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	oauthConnectionServiceMock.EXPECT().GetOAuthConnectionDetails(mock.Anything, mock.MatchedBy(func(req *models.GetOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" && req.Provider == "google"
	})).Return(nil, customerrors.ErrOAuthConnectionNotFound)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Get("/api/v1/projects/:project_id/oauth-connections/:provider", h.getOAuthConnectionDetails)

	req := httptest.NewRequest("GET", "/api/v1/projects/proj-123/oauth-connections/google", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	// ErrOAuthConnectionNotFound returns 404 Not Found
	assert.Equal(t, 404, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, customerrors.ErrOAuthConnectionNotFound.Code, parsed.Get("error.code").String())
}

func TestCreateProjectOAuthConnection_Success(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	now := time.Now()
	expectedResponse := &models.OAuthConnectionDetails{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: "custom",
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	oauthConnectionServiceMock.EXPECT().CreateOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.CreateOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google" &&
			req.CredentialType == "custom" &&
			req.ClientID == "client-1" &&
			req.ClientSecret == "secret-1" &&
			len(req.Scopes) == 2 &&
			req.Enabled != nil && *req.Enabled == true
	})).Return(expectedResponse, nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Post("/api/v1/projects/:project_id/oauth-connections/:provider", h.createOAuthConnection)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"credential_type": "custom",
		"client_id":       "client-1",
		"client_secret":   "secret-1",
		"scopes":          []string{"openid", "profile"},
		"enabled":         true,
	})

	req := httptest.NewRequest("POST", "/api/v1/projects/proj-123/oauth-connections/google", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "conn-1", parsed.Get("data.id").String())
	assert.Equal(t, "proj-123", parsed.Get("data.project_id").String())
	assert.Equal(t, "google", parsed.Get("data.provider").String())
	assert.Equal(t, "custom", parsed.Get("data.credential_type").String())
	assert.True(t, parsed.Get("data.enabled").Bool())
}

func TestCreateProjectOAuthConnection_ValidationError(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Post("/api/v1/projects/:project_id/oauth-connections/:provider", h.createOAuthConnection)

	// Missing required credential_type field
	requestBody, _ := json.Marshal(map[string]interface{}{})

	req := httptest.NewRequest("POST", "/api/v1/projects/proj-123/oauth-connections/google", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.NotEmpty(t, parsed.Get("error.message").String())
}

func TestCreateProjectOAuthConnection_ServiceError(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	oauthConnectionServiceMock.EXPECT().CreateOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.CreateOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google" &&
			req.CredentialType == "custom" &&
			req.ClientID == "client-1" &&
			req.ClientSecret == "secret-1" &&
			len(req.Scopes) == 1
	})).Return(nil, customerrors.ErrOAuthConnectionAlreadyExists)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Post("/api/v1/projects/:project_id/oauth-connections/:provider", h.createOAuthConnection)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"credential_type": "custom",
		"client_id":       "client-1",
		"client_secret":   "secret-1",
		"scopes":          []string{"openid"},
	})

	req := httptest.NewRequest("POST", "/api/v1/projects/proj-123/oauth-connections/google", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	// ErrOAuthConnectionAlreadyExists returns 409 Conflict
	assert.Equal(t, 409, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, customerrors.ErrOAuthConnectionAlreadyExists.Code, parsed.Get("error.code").String())
	assert.Equal(t, customerrors.ErrOAuthConnectionAlreadyExists.Msg, parsed.Get("error.message").String())
}

func TestUpdateProjectOAuthConnection_Success(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	now := time.Now()
	expectedResponse := &models.OAuthConnectionDetails{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: "custom",
		Enabled:        false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	oauthConnectionServiceMock.EXPECT().UpdateOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.UpdateOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google" &&
			req.CredentialType == "custom" &&
			req.ClientID == "client-2" &&
			req.ClientSecret == "secret-2" &&
			len(req.Scopes) == 3 &&
			req.Enabled != nil && *req.Enabled == false
	})).Return(expectedResponse, nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Put("/api/v1/projects/:project_id/oauth-connections/:provider", h.updateOAuthConnection)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"credential_type": "custom",
		"client_id":       "client-2",
		"client_secret":   "secret-2",
		"scopes":          []string{"openid", "profile", "email"},
		"enabled":         false,
	})

	req := httptest.NewRequest("PUT", "/api/v1/projects/proj-123/oauth-connections/google", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "conn-1", parsed.Get("data.id").String())
	assert.False(t, parsed.Get("data.enabled").Bool())
}

func TestUpdateProjectOAuthConnection_ServiceError(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	oauthConnectionServiceMock.EXPECT().UpdateOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.UpdateOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google" &&
			req.CredentialType == "custom" &&
			req.ClientID == "client-1" &&
			req.ClientSecret == "secret-1" &&
			len(req.Scopes) == 1
	})).Return(nil, customerrors.ErrOAuthConnectionNotFound)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Put("/api/v1/projects/:project_id/oauth-connections/:provider", h.updateOAuthConnection)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"credential_type": "custom",
		"client_id":       "client-1",
		"client_secret":   "secret-1",
		"scopes":          []string{"openid"},
	})

	req := httptest.NewRequest("PUT", "/api/v1/projects/proj-123/oauth-connections/google", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	res, err := app.Test(req)
	assert.NoError(t, err)
	// ErrOAuthConnectionNotFound returns 404 Not Found
	assert.Equal(t, 404, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, customerrors.ErrOAuthConnectionNotFound.Code, parsed.Get("error.code").String())
}

func TestDeleteProjectOAuthConnection_Success(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	oauthConnectionServiceMock.EXPECT().DeleteOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.DeleteOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google"
	})).Return(nil)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Delete("/api/v1/projects/:project_id/oauth-connections/:provider", h.deleteOAuthConnection)

	req := httptest.NewRequest("DELETE", "/api/v1/projects/proj-123/oauth-connections/google", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.True(t, parsed.Get("success").Bool())
}

func TestDeleteProjectOAuthConnection_ServiceError(t *testing.T) {
	app := fiber.New()
	oauthConnectionServiceMock := servicemocks.NewMockOAuthConnectionService(t)

	oauthConnectionServiceMock.EXPECT().DeleteOAuthConnection(mock.Anything, mock.MatchedBy(func(req *models.DeleteOAuthConnectionRequest) bool {
		return req.ProjectID == "proj-123" &&
			req.Provider == "google"
	})).Return(customerrors.ErrOAuthConnectionNotFound)

	deps := &Dependencies{
		OAuthConnectionService: oauthConnectionServiceMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)

	app.Delete("/api/v1/projects/:project_id/oauth-connections/:provider", h.deleteOAuthConnection)

	req := httptest.NewRequest("DELETE", "/api/v1/projects/proj-123/oauth-connections/google", nil)

	res, err := app.Test(req)
	assert.NoError(t, err)
	// ErrOAuthConnectionNotFound returns 404 Not Found
	assert.Equal(t, 404, res.StatusCode)

	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)

	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, customerrors.ErrOAuthConnectionNotFound.Code, parsed.Get("error.code").String())
}
