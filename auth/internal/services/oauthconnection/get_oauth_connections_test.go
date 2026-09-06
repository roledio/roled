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

func TestGetOAuthConnections_Success(t *testing.T) {
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

	connections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			ClientID:              "client-1",
			ClientSecretEncrypted: "secret-1",
			Scopes:                "openid profile email",
			Enabled:               true,
		},
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "proj-123").Return(connections, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionsRequest{
		ProjectID: "proj-123",
	}

	result, err := service.GetOAuthConnections(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "conn-1", result[0].ID)
	assert.Equal(t, "google", result[0].Provider)
}

func TestGetOAuthConnections_ProjectNotFound(t *testing.T) {
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

	service := NewOAuthConnectionService(&configs.DefaultConfig{}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionsRequest{
		ProjectID: "proj-123",
	}

	result, err := service.GetOAuthConnections(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestGetOAuthConnections_ProjectDBError(t *testing.T) {
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

	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(nil, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionsRequest{
		ProjectID: "proj-123",
	}

	result, err := service.GetOAuthConnections(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetOAuthConnections_ConnectionsDBError(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "proj-123").Return(nil, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionsRequest{
		ProjectID: "proj-123",
	}

	result, err := service.GetOAuthConnections(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetOAuthConnections_EmptyList(t *testing.T) {
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

	connections := []entities.OAuthConnection{}
	mockOAuthConnectionRepo.EXPECT().FindByProjectID(ctx, "proj-123").Return(connections, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionsRequest{
		ProjectID: "proj-123",
	}

	result, err := service.GetOAuthConnections(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}
