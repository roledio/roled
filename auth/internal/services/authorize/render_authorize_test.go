package authorize

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizeService_RenderAuthorize_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

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
	mockClientRepo := repositorymocks.NewMockClientRepository(t)

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
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

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
	mockClientRepo := repositorymocks.NewMockClientRepository(t)

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
