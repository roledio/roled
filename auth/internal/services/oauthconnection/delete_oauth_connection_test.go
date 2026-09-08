package oauthconnection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
)

func TestDeleteOAuthConnection_Success(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().Delete(ctx, "proj-123", "google").Return(1, nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.DeleteOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	err := service.DeleteOAuthConnection(ctx, req)

	assert.NoError(t, err)
}

func TestDeleteOAuthConnection_ProjectNotFound(t *testing.T) {
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

	req := &models.DeleteOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	err := service.DeleteOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
}

func TestDeleteOAuthConnection_ConnectionNotFound(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().Delete(ctx, "proj-123", "google").Return(0, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.DeleteOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	err := service.DeleteOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
}

func TestDeleteOAuthConnection_DBError(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().Delete(ctx, "proj-123", "google").Return(0, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.DeleteOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	err := service.DeleteOAuthConnection(ctx, req)

	assert.Error(t, err)
}
