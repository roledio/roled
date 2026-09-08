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

func TestUpdateOAuthConnection_Success_Custom(t *testing.T) {
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

	existingConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              stringPtr("old-client"),
		ClientSecretEncrypted: stringPtr("old-secret"),
		Scopes:                stringPtr("openid"),
		Enabled:               true,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(1, nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "new-client",
		ClientSecret:   "new-secret",
		Scopes:         []string{"openid", "profile", "email"},
		Enabled:        boolPtr(false),
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeCustom, result.CredentialType)
	assert.False(t, result.Enabled)
}

func TestUpdateOAuthConnection_Success_Default(t *testing.T) {
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

	existingConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              stringPtr("old-client"),
		ClientSecretEncrypted: stringPtr("old-secret"),
		Scopes:                stringPtr("openid"),
		Enabled:               true,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(1, nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeDefault,
		Enabled:        boolPtr(true),
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeDefault, result.CredentialType)
	assert.True(t, result.Enabled)
}

func TestUpdateOAuthConnection_ProjectNotFound(t *testing.T) {
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

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_ConnectionNotFound(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(nil, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_UpdateFails(t *testing.T) {
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

	existingConnection := &entities.OAuthConnection{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(0, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_NoRowsAffected(t *testing.T) {
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

	existingConnection := &entities.OAuthConnection{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(0, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func stringPtr(s string) *string {
	return &s
}
