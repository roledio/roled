package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	"github.com/roledio/roled/auth/internal/repositories/redis"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
)

func TestCachedOAuthConnectionRepository_FindByProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	expectedConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Run(func(ctx context.Context, key string, dest any) {
			conns := dest.(*[]entities.OAuthConnection)
			*conns = expectedConnections
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "conn-1", result[0].ID)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	individualCacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "google")

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	mockRedis.EXPECT().
		SetData(ctx, mainCacheKey, dbConnections, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey, dbConnections[0], 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "conn-1", result[0].ID)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_CacheMissMultipleConnections(t *testing.T) {
	ctx := context.Background()
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
		{
			ID:                    "conn-2",
			ProjectID:             "proj-123",
			Provider:              "github",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-2"),
			ClientSecretEncrypted: pointy.String("secret-2"),
			Scopes:                pointy.String("read:user user:email"),
			Enabled:               true,
		},
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	individualCacheKey1 := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "google")
	individualCacheKey2 := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "github")

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	mockRedis.EXPECT().
		SetData(ctx, mainCacheKey, dbConnections, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey1, dbConnections[0], 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey2, dbConnections[1], 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_CacheError(t *testing.T) {
	ctx := context.Background()
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	individualCacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "google")

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, assert.AnError)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	mockRedis.EXPECT().
		SetData(ctx, mainCacheKey, dbConnections, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey, dbConnections[0], 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_DBError(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(nil, assert.AnError)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_NilRedis(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, nil, 24*time.Hour)

	projectID := "proj-123"
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)
	expectedConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Run(func(ctx context.Context, key string, dest any) {
			conn := dest.(*entities.OAuthConnection)
			*conn = *expectedConnection
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "google", result.Provider)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, projectID, provider).
		Return(dbConnection, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbConnection, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_CacheError(t *testing.T) {
	ctx := context.Background()
	dbConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Return(false, assert.AnError)

	mockDBRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, projectID, provider).
		Return(dbConnection, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbConnection, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_DBError(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, projectID, provider).
		Return(nil, assert.AnError)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, projectID, provider).
		Return(nil, nil)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCachedOAuthConnectionRepository_FindByProjectIDAndProvider_CacheMissSetError(t *testing.T) {
	ctx := context.Background()
	dbConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	provider := "google"
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectIDAndProvider(ctx, projectID, provider).
		Return(dbConnection, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbConnection, 24*time.Hour).
		Return(assert.AnError)

	result, err := cachedRepo.FindByProjectIDAndProvider(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
}

func TestCachedOAuthConnectionRepository_Create(t *testing.T) {
	ctx := context.Background()
	connection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Create(ctx, connection).
		Return(nil)

	err := cachedRepo.Create(ctx, connection)

	assert.NoError(t, err)
}

func TestCachedOAuthConnectionRepository_Create_Error(t *testing.T) {
	ctx := context.Background()
	connection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Create(ctx, connection).
		Return(assert.AnError)

	err := cachedRepo.Create(ctx, connection)

	assert.Error(t, err)
}

func TestCachedOAuthConnectionRepository_Update(t *testing.T) {
	ctx := context.Background()
	connection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, connection).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, connection)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedOAuthConnectionRepository_Update_Error(t *testing.T) {
	ctx := context.Background()
	connection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              pointy.String("client-1"),
		ClientSecretEncrypted: pointy.String("secret-1"),
		Scopes:                pointy.String("openid profile email"),
		Enabled:               true,
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, connection).
		Return(0, assert.AnError)

	affected, err := cachedRepo.Update(ctx, connection)

	assert.Error(t, err)
	assert.Equal(t, 0, affected)
}

func TestCachedOAuthConnectionRepository_Delete(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"
	provider := "google"

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Delete(ctx, projectID, provider).
		Return(1, nil)

	affected, err := cachedRepo.Delete(ctx, projectID, provider)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedOAuthConnectionRepository_Delete_Error(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"
	provider := "google"

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Delete(ctx, projectID, provider).
		Return(0, assert.AnError)

	affected, err := cachedRepo.Delete(ctx, projectID, provider)

	assert.Error(t, err)
	assert.Equal(t, 0, affected)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_SetError(t *testing.T) {
	ctx := context.Background()
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	individualCacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "google")

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	mockRedis.EXPECT().
		SetData(ctx, mainCacheKey, dbConnections, 24*time.Hour).
		Return(assert.AnError)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey, dbConnections[0], 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
}

func TestCachedOAuthConnectionRepository_FindByProjectID_IndividualSetError(t *testing.T) {
	ctx := context.Background()
	dbConnections := []entities.OAuthConnection{
		{
			ID:                    "conn-1",
			ProjectID:             "proj-123",
			Provider:              "google",
			CredentialType:        constants.OAuthCredentialTypeCustom,
			ClientID:              pointy.String("client-1"),
			ClientSecretEncrypted: pointy.String("secret-1"),
			Scopes:                pointy.String("openid profile email"),
			Enabled:               true,
		},
	}

	mockDBRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	individualCacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, "google")

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*[]entities.OAuthConnection")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	mockRedis.EXPECT().
		SetData(ctx, mainCacheKey, dbConnections, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, individualCacheKey, dbConnections[0], 24*time.Hour).
		Return(assert.AnError)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
}
