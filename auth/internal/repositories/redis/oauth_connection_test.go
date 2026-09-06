package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedOAuthConnectionRepository_FindByProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	expectedConnections := []entities.OAuthConnection{
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
			ClientID:              "client-1",
			ClientSecretEncrypted: "secret-1",
			Scopes:                "openid profile email",
			Enabled:               true,
		},
	}

	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

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
			ClientID:              "client-1",
			ClientSecretEncrypted: "secret-1",
			Scopes:                "openid profile email",
			Enabled:               true,
		},
		{
			ID:                    "conn-2",
			ProjectID:             "proj-123",
			Provider:              "github",
			ClientID:              "client-2",
			ClientSecretEncrypted: "secret-2",
			Scopes:                "read:user user:email",
			Enabled:               true,
		},
	}

	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

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
			ClientID:              "client-1",
			ClientSecretEncrypted: "secret-1",
			Scopes:                "openid profile email",
			Enabled:               true,
		},
	}

	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

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
	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

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
	mockDBRepo := repositorymocks.NewMockOAuthConnectionRepository(t)

	cachedRepo := redis.NewOAuthConnectionRepository(mockDBRepo, nil, 24*time.Hour)

	projectID := "proj-123"
	dbConnections := []entities.OAuthConnection{
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

	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbConnections, nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
}
