package accesstoken

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"
)

func createCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func TestAccessTokenService_ExchangeToken_AuthorizationCode_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRefreshTokenRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)
	mockRegistry.EXPECT().RefreshTokenRepository().Return(mockRefreshTokenRepo)
	mockRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)

	// Mock Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

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

	// Mock auth code found
	expiresAt := time.Now().Add(time.Hour)
	authCode := &entities.AuthCode{
		ID:                  "auth-code-id",
		AccountID:           "test-account",
		ProjectID:           "test-project",
		ClientID:            "test-client",
		UserID:              pointy.String("test-user"),
		CodeHash:            "hash",
		CodeChallenge:       createCodeChallenge("test-verifier"),
		CodeChallengeMethod: "S256",
		RedirectURI:         "http://example.com/callback",
		State:               pointy.String("test-state"),
		ExpiresAt:           expiresAt,
		UsedAt:              nil,
	}
	mockAuthCodeRepo.EXPECT().FindByClientIDAndCodeHash(ctx, "test-client", mock.AnythingOfType("string")).Return(authCode, nil)

	// Mock update used auth code
	mockAuthCodeRepo.EXPECT().UpdateUsedAuthCode(ctx, authCode).Return(1, nil)

	// Mock create new refresh token
	mockRefreshTokenRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.RefreshToken")).Return(nil)

	// Mock create new access token
	mockAccessTokenRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AccessToken")).Return(nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.SigningKey = "test-signing-key"
	defaultConfig.JWT.ExpiryDuration = "1h"
	defaultConfig.JWT.RefreshTokenExpiryDuration = "7d"
	defaultConfig.BaseURL = "http://test.com"
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request
	req := &models.ExchangeTokenRequest{
		GrantType:         constants.GrantTypeAuthorizationCode,
		ClientID:          "test-client",
		AuthorizationCode: "test-auth-code",
		RedirectURI:       "http://example.com/callback",
		CodeVerifier:      "test-verifier",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "bearer", result.TokenType)
	assert.Greater(t, result.ExpiresIn, 0)
	assert.Greater(t, result.RefreshTokenExpiresIn, 0)
}

func TestAccessTokenService_ExchangeToken_AuthorizationCode_InvalidCode(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)

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

	// Mock auth code not found
	mockAuthCodeRepo.EXPECT().FindByClientIDAndCodeHash(ctx, "test-client", mock.AnythingOfType("string")).Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Test request
	req := &models.ExchangeTokenRequest{
		GrantType:         constants.GrantTypeAuthorizationCode,
		ClientID:          "test-client",
		AuthorizationCode: "invalid-auth-code",
		RedirectURI:       "http://example.com/callback",
		CodeVerifier:      "test-verifier",
	}

	// Execute
	result, err := service.ExchangeToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidAuthorizationCode, err)
	assert.Nil(t, result)
}
