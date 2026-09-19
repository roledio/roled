package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	"github.com/roledio/roled/auth/internal/repositories/redis"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
)

func TestCachedUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	provider := "github"
	providerUserID := "github_user_123"
	projectID := "proj_123"
	cacheKey := rediskeys.UserIdentityByProviderAndProviderUserIDAndProjectID(provider, providerUserID, projectID)
	expectedUserIdentity := &entities.UserIdentity{
		ID:             "uid_123",
		UserID:         "user_456",
		Provider:       provider,
		ProviderUserID: providerUserID,
		ProjectID:      projectID,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.UserIdentity")).
		Run(func(ctx context.Context, key string, dest any) {
			uid := dest.(*entities.UserIdentity)
			*uid = *expectedUserIdentity
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUserIdentity.ID, result.ID)
	assert.Equal(t, expectedUserIdentity.Provider, result.Provider)
	assert.Equal(t, expectedUserIdentity.ProviderUserID, result.ProviderUserID)
}

func TestCachedUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbUserIdentity := &entities.UserIdentity{
		ID:             "uid_123",
		UserID:         "user_456",
		Provider:       "github",
		ProviderUserID: "github_user_123",
		ProjectID:      "proj_123",
	}

	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	provider := "github"
	providerUserID := "github_user_123"
	projectID := "proj_123"
	cacheKey := rediskeys.UserIdentityByProviderAndProviderUserIDAndProjectID(provider, providerUserID, projectID)
	cacheKeyByID := rediskeys.UserIdentityByID(dbUserIdentity.ID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.UserIdentity")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID).
		Return(dbUserIdentity, nil)

	// Mock Redis SetData saving to cache with both keys
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbUserIdentity, 24*time.Hour).
		Return(nil)
	mockRedis.EXPECT().
		SetData(ctx, cacheKeyByID, dbUserIdentity, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbUserIdentity.ID, result.ID)
}

func TestCachedUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID_CacheMiss_NotFound(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	provider := "github"
	providerUserID := "github_user_123"
	projectID := "proj_123"
	cacheKey := rediskeys.UserIdentityByProviderAndProviderUserIDAndProjectID(provider, providerUserID, projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.UserIdentity")).
		Return(false, nil)

	// Mock DB query returning nil (not found)
	mockDBRepo.EXPECT().
		FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID).
		Return(nil, nil)

	result, err := cachedRepo.FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestCachedUserIdentityRepository_FindByProviderAndProviderUserIDAndProjectID_CacheMiss_DBError(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	provider := "github"
	providerUserID := "github_user_123"
	projectID := "proj_123"
	cacheKey := rediskeys.UserIdentityByProviderAndProviderUserIDAndProjectID(provider, providerUserID, projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.UserIdentity")).
		Return(false, nil)

	// Mock DB query returning error
	mockDBRepo.EXPECT().
		FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID).
		Return(nil, assert.AnError)

	result, err := cachedRepo.FindByProviderAndProviderUserIDAndProjectID(ctx, provider, providerUserID, projectID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCachedUserIdentityRepository_Create(t *testing.T) {
	ctx := context.Background()
	userIdentity := &entities.UserIdentity{
		ID:             "uid_123",
		UserID:         "user_456",
		Provider:       "github",
		ProviderUserID: "github_user_123",
		ProjectID:      "proj_123",
	}

	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, userIdentity).
		Return(nil)

	err := cachedRepo.Create(ctx, userIdentity)

	assert.NoError(t, err)
}

func TestCachedUserIdentityRepository_FindByUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user_123"
	expectedUserIdentities := []*entities.UserIdentity{
		{
			ID:             "uid_1",
			UserID:         userID,
			Provider:       "github",
			ProviderUserID: "github_user_1",
			ProjectID:      "proj_123",
		},
		{
			ID:             "uid_2",
			UserID:         userID,
			Provider:       "google",
			ProviderUserID: "google_user_1",
			ProjectID:      "proj_123",
		},
	}

	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindByUserID (cache should not be used for list queries)
	mockDBRepo.EXPECT().
		FindByUserID(ctx, userID).
		Return(expectedUserIdentities, nil)

	result, err := cachedRepo.FindByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedUserIdentities), len(result))
}

func TestCachedUserIdentityRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	userIdentityID := "uid_123"

	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)
	mockRedis := redismocks.NewMockService(t)

	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByID(ctx, userIdentityID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByID(ctx, userIdentityID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserIdentityRepository_NilRedis(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := interfacemocks.NewMockUserIdentityRepository(t)

	// Create cached repo with nil redis - should return the underlying repo directly
	cachedRepo := redis.NewUserIdentityRepository(mockDBRepo, nil, 24*time.Hour)

	userIdentity := &entities.UserIdentity{
		ID:             "uid_123",
		UserID:         "user_456",
		Provider:       "github",
		ProviderUserID: "github_user_123",
		ProjectID:      "proj_123",
	}

	// Test Create - should delegate to DB repo
	mockDBRepo.EXPECT().
		Create(ctx, userIdentity).
		Return(nil)

	err := cachedRepo.Create(ctx, userIdentity)
	assert.NoError(t, err)

	// Test FindByProviderAndProviderUserIDAndProjectID - should delegate to DB repo
	mockDBRepo.EXPECT().
		FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_123", "proj_123").
		Return(userIdentity, nil)

	result, err := cachedRepo.FindByProviderAndProviderUserIDAndProjectID(ctx, "github", "github_user_123", "proj_123")
	assert.NoError(t, err)
	assert.Equal(t, userIdentity.ID, result.ID)
}
