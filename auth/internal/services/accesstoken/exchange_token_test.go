package accesstoken

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
)

func TestAccessTokenService_ExchangeToken_InvalidGrantType(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request with invalid grant type
	req := &models.ExchangeTokenRequest{
		GrantType: "invalid_grant",
		ClientID:  "test-client",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidGrantType, err)
	assert.Nil(t, result)
}
