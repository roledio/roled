package authorize

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizeService_RenderAuthorize_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock oauth connection not found
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "authorization_code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project, result.Project)
	assert.Equal(t, projectSetting, result.ProjectSetting)
}

func TestAuthorizeService_RenderAuthorize_ClientNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)

	// Mock registry to return client repo
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock client not found
	mockClientRepo.EXPECT().FindByID(ctx, "invalid-client").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "invalid-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientID, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_InvalidRedirectURI(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI not found
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://invalid.com/callback").Return(nil, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request with invalid redirect URL
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://invalid.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidRedirectURI, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_DatabaseError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)

	// Mock registry to return client repo
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock database error
	dbErr := assert.AnError
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_ClientNotActive(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)

	// Mock registry to return client repo
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock client found but not active
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  false,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrClientNotActive, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_AccountNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account not found
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientID, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_AccountNotActive(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account found but not active
	account := &entities.Account{
		ID:       "test-account",
		IsActive: false,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientID, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock project not found
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientID, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_ProjectNotActive(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock project found but not active
	project := &entities.Project{
		ID:       "test-project",
		IsActive: false,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotActive, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_WithForgotPassword(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found with forgot password enabled
	projectSetting := &entities.ProjectSetting{
		ProjectID:               "test-project",
		IsForgotPasswordEnabled: true,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock oauth connection not found
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project, result.Project)
	assert.Equal(t, projectSetting, result.ProjectSetting)
	assert.NotEmpty(t, result.ForgotPasswordURLPath)
	assert.Contains(t, result.ForgotPasswordURLPath, "/password/forgot")
	assert.Contains(t, result.ForgotPasswordURLPath, "client_id=test-client")
	assert.Contains(t, result.ForgotPasswordURLPath, "redirect_uri=http%3A%2F%2Fexample.com%2Fcallback")
}

func TestAuthorizeService_RenderAuthorize_WithOAuthConnections(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock oauth connections found
	conns := []entities.OAuthConnection{
		{
			ID:        "google-conn",
			ProjectID: "test-project",
			Provider:  "google",
			Enabled:   true,
		},
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(conns, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		ResponseType:        "code",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
		State:               "test-state",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project, result.Project)
	assert.Equal(t, projectSetting, result.ProjectSetting)
	assert.Len(t, result.OAuthConnections, 1)
	assert.Equal(t, "google", result.OAuthConnections[0].Provider)
	assert.NotEmpty(t, result.OAuthConnections[0].URLPath)
	assert.Contains(t, result.OAuthConnections[0].URLPath, "/oauth/google")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "client_id=test-client")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "redirect_uri=http%3A%2F%2Fexample.com%2Fcallback")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "response_type=code")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "code_challenge=test-challenge")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "code_challenge_method=S256")
	assert.Contains(t, result.OAuthConnections[0].URLPath, "state=test-state")
}

func TestAuthorizeService_RenderAuthorize_OAuthConnectionDisabled(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock oauth connections found but disabled
	conns := []entities.OAuthConnection{
		{
			ID:        "google-conn",
			ProjectID: "test-project",
			Provider:  "google",
			Enabled:   false,
		},
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(conns, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project, result.Project)
	assert.Equal(t, projectSetting, result.ProjectSetting)
	assert.Len(t, result.OAuthConnections, 0)
}

func TestAuthorizeService_RenderAuthorize_OAuthConnectionError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock oauth connection repository error
	dbErr := assert.AnError
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project, result.Project)
	assert.Equal(t, projectSetting, result.ProjectSetting)
	assert.Len(t, result.OAuthConnections, 0)
}

func TestAuthorizeService_RenderAuthorize_ProjectSettingError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := interfacemocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock project setting repository error
	dbErr := assert.AnError
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_AccountRepositoryError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account repository error
	dbErr := assert.AnError
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_ProjectRepositoryError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock project repository error
	dbErr := assert.AnError
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}

func TestAuthorizeService_RenderAuthorize_RedirectURIRepositoryError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := interfacemocks.NewMockClientRepository(t)
	mockAccountRepo := interfacemocks.NewMockAccountRepository(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockRedirectURIRepo := interfacemocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI repository error
	dbErr := assert.AnError
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(nil, dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.RenderAuthorizeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://example.com/callback",
		ResponseType: "code",
	}

	// Execute
	result, err := service.RenderAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}
