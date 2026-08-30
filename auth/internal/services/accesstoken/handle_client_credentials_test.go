package accesstoken

import (
	"context"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createEncryptedSecret(plainSecret, masterKey string) string {
	key, _ := encryptionutil.DeriveKey([]byte(masterKey), pkgconstants.KeyPurposeClientSecret)
	encrypted, _ := encryptionutil.EncryptAES(plainSecret, key, pkgconstants.KeyPurposeClientSecret)
	return encrypted
}

func TestAccessTokenService_ExchangeToken_ClientCredentials_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)

	// Mock client found
	client := &entities.Client{
		ID:              "test-client",
		ProjectID:       "test-project",
		AccountID:       "test-account",
		IsActive:        true,
		SecretEncrypted: createEncryptedSecret("test-secret", "test-master-key"),
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock access token create success
	mockAccessTokenRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AccessToken")).Return(nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.EncryptionMasterKey = "test-master-key"
	defaultConfig.JWT.SigningKey = "test-signing-key"
	defaultConfig.JWT.ExpiryDuration = "1h"
	defaultConfig.BaseURL = "http://test.com"
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request
	req := &models.ExchangeTokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "test-client",
		ClientSecret: "test-secret", // Assuming decryption works
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "bearer", result.TokenType)
	assert.Greater(t, result.ExpiresIn, 0)
}

func TestAccessTokenService_ExchangeToken_ClientCredentials_InvalidClient(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock client not found
	mockClientRepo.EXPECT().FindByID(ctx, "invalid-client").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request
	req := &models.ExchangeTokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "invalid-client",
		ClientSecret: "test-secret",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientID, err)
	assert.Nil(t, result)
}

func TestAccessTokenService_ExchangeToken_ClientCredentials_InactiveClient(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)

	// Mock inactive client
	client := &entities.Client{
		ID:       "test-client",
		IsActive: false,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request
	req := &models.ExchangeTokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrClientNotActive, err)
	assert.Nil(t, result)
}

func TestAccessTokenService_ExchangeToken_ClientCredentials_InvalidSecret(t *testing.T) {
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
		ID:              "test-client",
		ProjectID:       "test-project",
		IsActive:        true,
		SecretEncrypted: createEncryptedSecret("correct-secret", "test-master-key"), // Different secret
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
	defaultConfig.EncryptionMasterKey = "test-master-key"
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request with wrong secret
	req := &models.ExchangeTokenRequest{
		GrantType:    constants.GrantTypeClientCredentials,
		ClientID:     "test-client",
		ClientSecret: "wrong-secret",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidClientCredentials, err)
	assert.Nil(t, result)
}
