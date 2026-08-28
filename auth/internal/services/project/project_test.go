package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/pkg/errors"

	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
)

func TestProjectService_GetConsoleConfig_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock project found
	project := &entities.Project{
		ID:       "console-project-id",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(project, nil)

	// Mock client found
	client := &entities.Client{
		ID:        "console-client-id",
		ProjectID: "console-project-id",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByProjectIDAndIsDefault(ctx, "console-project-id", true).Return(client, nil)

	// Create service
	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)

	// Execute
	result, err := service.GetConsoleConfig(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "console-client-id", result.ClientID)
}

func TestProjectService_GetConsoleConfig_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	// Mock registry to return project repo
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project not found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(nil, nil)

	// Create service
	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)

	// Execute
	result, err := service.GetConsoleConfig(ctx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrSystemError.Msg, err.Error())
	assert.Nil(t, result)
}

func TestProjectService_GetConsoleConfig_DatabaseError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	// Mock registry to return project repo
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock database error
	dbErr := assert.AnError
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(nil, dbErr)

	// Create service
	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)

	// Execute
	result, err := service.GetConsoleConfig(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}
