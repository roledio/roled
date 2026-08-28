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

func TestCachedProjectSettingRepository_FindByProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectSettingRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.ProjectSettingByProjectID(projectID)
	expectedSetting := &entities.ProjectSetting{
		ID:        "ps-123",
		ProjectID: projectID,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.ProjectSetting")).
		Run(func(ctx context.Context, key string, dest any) {
			ps := dest.(*entities.ProjectSetting)
			*ps = *expectedSetting
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSetting.ID, result.ID)
	assert.Equal(t, expectedSetting.ProjectID, result.ProjectID)
}

func TestCachedProjectSettingRepository_FindByProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbSetting := &entities.ProjectSetting{
		ID:        "ps-123",
		ProjectID: "proj-123",
	}

	mockDBRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectSettingRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.ProjectSettingByProjectID(projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.ProjectSetting")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectID(ctx, projectID).
		Return(dbSetting, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbSetting, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbSetting.ID, result.ID)
}

func TestCachedProjectSettingRepository_Create(t *testing.T) {
	ctx := context.Background()
	setting := &entities.ProjectSetting{
		ID:        "ps-123",
		ProjectID: "proj-123",
	}

	mockDBRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectSettingRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, setting).
		Return(nil)

	err := cachedRepo.Create(ctx, setting)

	assert.NoError(t, err)
}

func TestCachedProjectSettingRepository_Update(t *testing.T) {
	ctx := context.Background()
	setting := &entities.ProjectSetting{
		ID:        "ps-123",
		ProjectID: "proj-123",
	}

	mockDBRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectSettingRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, setting).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, setting)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}
