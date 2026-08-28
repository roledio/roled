package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/entities"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/roledio/roled/internal/repositories/redis"
)

func TestCachedRedirectURIRepository_FindByProjectIDAndRedirectURI_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	redirectURI := "https://example.com/callback"
	cacheKey := rediskeys.RedirectURIByProjectIDAndURI(projectID, redirectURI)
	loginURL := "https://example.com/login"
	expectedURI := &entities.RedirectURI{
		ProjectID:   projectID,
		RedirectURI: redirectURI,
		LoginURL:    &loginURL,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.RedirectURI")).
		Run(func(ctx context.Context, key string, dest any) {
			uri := dest.(*entities.RedirectURI)
			*uri = *expectedURI
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndRedirectURI(ctx, projectID, redirectURI)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURI.RedirectURI, result.RedirectURI)
}

func TestCachedRedirectURIRepository_FindByProjectIDAndRedirectURI_CacheMiss(t *testing.T) {
	ctx := context.Background()
	loginURL := "https://example.com/login"
	dbURI := &entities.RedirectURI{
		ProjectID:   "proj-123",
		RedirectURI: "https://example.com/callback",
		LoginURL:    &loginURL,
	}

	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	redirectURI := "https://example.com/callback"
	cacheKey := rediskeys.RedirectURIByProjectIDAndURI(projectID, redirectURI)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.RedirectURI")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectIDAndRedirectURI(ctx, projectID, redirectURI).
		Return(dbURI, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbURI, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndRedirectURI(ctx, projectID, redirectURI)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbURI.RedirectURI, result.RedirectURI)
}

func TestCachedRedirectURIRepository_Create(t *testing.T) {
	ctx := context.Background()
	loginURL := "https://example.com/login"
	uris := []entities.RedirectURI{
		{ProjectID: "proj-123", RedirectURI: "https://example.com/callback1", LoginURL: &loginURL},
		{ProjectID: "proj-123", RedirectURI: "https://example.com/callback2", LoginURL: &loginURL},
	}

	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, uris).
		Return(nil)

	err := cachedRepo.Create(ctx, uris)

	assert.NoError(t, err)
}

func TestCachedRedirectURIRepository_DeleteByProjectID(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"

	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByProjectID(ctx, projectID).
		Return(2, nil)

	affected, err := cachedRepo.DeleteByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.Equal(t, 2, affected)
}

func TestCachedRedirectURIRepository_FindByProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	redirectURI := "https://example.com/callback"
	cacheKey := rediskeys.RedirectURIsByProjectID(projectID)
	loginURL := "https://example.com/login"
	expectedURI := entities.RedirectURI{
		ProjectID:   projectID,
		RedirectURI: redirectURI,
		LoginURL:    &loginURL,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*[]entities.RedirectURI")).
		Run(func(ctx context.Context, key string, dest any) {
			uris := dest.(*[]entities.RedirectURI)
			*uris = []entities.RedirectURI{expectedURI}
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedURI.RedirectURI, result[0].RedirectURI)
}

func TestCachedRedirectURIRepository_FindByProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	loginURL := "https://example.com/login"
	dbURI := entities.RedirectURI{
		ProjectID:   "proj-123",
		RedirectURI: "https://example.com/callback",
		LoginURL:    &loginURL,
	}

	mockDBRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRedirectURIRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.RedirectURIsByProjectID(projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*[]entities.RedirectURI")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return([]entities.RedirectURI{dbURI}, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, []entities.RedirectURI{dbURI}, 24*time.Hour).
		Return(nil)

	cacheKey = rediskeys.RedirectURIByProjectIDAndURI(projectID, dbURI.RedirectURI)
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbURI, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbURI.RedirectURI, result[0].RedirectURI)
}
