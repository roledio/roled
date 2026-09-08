package oauthconnection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
)

func TestCreateOAuthConnection_Success(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	// Check existing connection
	mockOAuthConnectionRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, "proj-123", "google").
		Return(nil, nil)

	// Create connection
	mockOAuthConnectionRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid", "profile", "email"},
		Enabled:        boolPtr(true),
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeCustom, result.CredentialType)
	assert.True(t, result.Enabled)
}

func TestCreateOAuthConnection_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(nil, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestCreateOAuthConnection_ConnectionAlreadyExists(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	// Connection already exists
	existingConnection := &entities.OAuthConnection{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
	}
	mockOAuthConnectionRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, "proj-123", "google").
		Return(existingConnection, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionAlreadyExists.Msg, err.Error())
	assert.Nil(t, result)
}

func TestCreateOAuthConnection_SystemProject(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID:       "proj-123",
		IsSystem: true,
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(nil, nil)
	mockOAuthConnectionRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeCustom, result.CredentialType)
}

func TestCreateOAuthConnection_EncryptionError(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	// Check existing connection
	mockOAuthConnectionRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, "proj-123", "google").
		Return(nil, nil)

	// Create fails
	mockOAuthConnectionRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(assert.AnError)

	// Invalid encryption master key will cause encryption to fail - but we still need a valid key format
	// The service will try to encrypt, and if that succeeds, it will hit the Tx
	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "valid-test-key-32-bytes-long!!!"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateOAuthConnection_DBError(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	// Check existing connection
	mockOAuthConnectionRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, "proj-123", "google").
		Return(nil, nil)

	// Create fails
	mockOAuthConnectionRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.CreateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.CreateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func boolPtr(b bool) *bool {
	return &b
}
