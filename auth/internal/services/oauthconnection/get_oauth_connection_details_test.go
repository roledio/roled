package oauthconnection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
)

func TestGetOAuthConnection_Success(t *testing.T) {
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

	secretKeyPlain := "abc123"
	encryptionKey := "test-key"
	purpose := constants.KeyPurposeOAuthClientSecret
	derivedKey, _ := encryptionutil.DeriveKey([]byte(encryptionKey), purpose)
	secretKeyEncrypted, _ := encryptionutil.EncryptAES(secretKeyPlain, derivedKey, purpose)

	now := time.Now()
	existingConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              new("client-id"),
		ClientSecretEncrypted: &secretKeyEncrypted,
		Scopes:                new("openid email profile"),
		Enabled:               true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: encryptionKey}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	result, err := service.GetOAuthConnectionDetails(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeCustom, result.CredentialType)
	assert.Equal(t, "client-id", *result.ClientID)
	assert.Equal(t, secretKeyPlain, *result.ClientSecret)
	assert.Equal(t, []string{"openid", "email", "profile"}, result.Scopes)
	assert.True(t, result.Enabled)
}

func TestGetOAuthConnection_ProjectNotFound(t *testing.T) {
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

	req := &models.GetOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	result, err := service.GetOAuthConnectionDetails(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestGetOAuthConnection_ConnectionNotFound(t *testing.T) {
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

	req := &models.GetOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	result, err := service.GetOAuthConnectionDetails(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestGetOAuthConnection_DBError(t *testing.T) {
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

	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(nil, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.GetOAuthConnectionRequest{
		ProjectID: "proj-123",
		Provider:  "google",
	}

	result, err := service.GetOAuthConnectionDetails(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}
