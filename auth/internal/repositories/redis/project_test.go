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
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedProjectRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.ProjectByID(projectID)
	expectedProject := &entities.Project{
		ID:       "proj-123",
		Name:     "Test Project",
		IsSystem: false,
		IsActive: true,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Project")).
		Run(func(ctx context.Context, key string, dest any) {
			pr := dest.(*entities.Project)
			*pr = *expectedProject
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedProject.ID, result.ID)
}

func TestCachedProjectRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbProject := &entities.Project{
		ID:       "proj-123",
		Name:     "Test Project",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	mainCacheKey := rediskeys.ProjectByID(projectID)
	setCacheKeys := []string{
		mainCacheKey,
		rediskeys.ProjectByIDAndAccountID(dbProject.ID, dbProject.AccountID),
	}

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*entities.Project")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByID(ctx, projectID).
		Return(dbProject, nil)

	for _, cacheKey := range setCacheKeys {
		mockRedis.EXPECT().
			SetData(ctx, cacheKey, dbProject, 24*time.Hour).
			Return(nil)
	}

	result, err := cachedRepo.FindByID(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbProject.ID, result.ID)
}

func TestCachedProjectRepository_FindByIDAndAccountID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	accountID := "acc-123"
	cacheKey := rediskeys.ProjectByIDAndAccountID(projectID, accountID)
	expectedProject := &entities.Project{
		ID:        "proj-123",
		AccountID: "acc-123",
		Name:      "Test Project",
		IsSystem:  false,
		IsActive:  true,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Project")).
		Run(func(ctx context.Context, key string, dest any) {
			pr := dest.(*entities.Project)
			*pr = *expectedProject
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByIDAndAccountID(ctx, projectID, accountID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedProject.ID, result.ID)
}

func TestCachedProjectRepository_FindByIDAndAccountID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbProject := &entities.Project{
		ID:        "proj-123",
		AccountID: "acc-123",
		Name:      "Test Project",
		IsSystem:  false,
		IsActive:  true,
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	accountID := "acc-123"
	mainCacheKey := rediskeys.ProjectByIDAndAccountID(projectID, accountID)
	setCacheKeys := []string{
		mainCacheKey,
		rediskeys.ProjectByID(projectID),
	}

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*entities.Project")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByIDAndAccountID(ctx, projectID, accountID).
		Return(dbProject, nil)

	for _, cacheKey := range setCacheKeys {
		mockRedis.EXPECT().
			SetData(ctx, cacheKey, dbProject, 24*time.Hour).
			Return(nil)
	}

	result, err := cachedRepo.FindByIDAndAccountID(ctx, projectID, accountID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbProject.ID, result.ID)
}

func TestCachedProjectRepository_FindSystem_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	expectedProject := &entities.Project{
		ID:       "system",
		Name:     "System Project",
		IsSystem: true,
		IsActive: true,
	}
	cacheKey := rediskeys.ProjectByIsSystem(true)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Project")).
		Run(func(ctx context.Context, key string, dest any) {
			pr := dest.(*entities.Project)
			*pr = *expectedProject
		}).
		Return(true, nil)

	result, err := cachedRepo.FindSystem(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "system", result.ID)
	assert.True(t, result.IsSystem)
}

func TestCachedProjectRepository_FindSystem_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbProject := &entities.Project{
		ID:       "system",
		Name:     "System Project",
		IsSystem: true,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mainCacheKey := rediskeys.ProjectByIsSystem(true)
	setCacheKeys := []string{
		mainCacheKey,
		rediskeys.ProjectByID(dbProject.ID),
		rediskeys.ProjectByIDAndAccountID(dbProject.ID, dbProject.AccountID),
	}

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*entities.Project")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindSystem(ctx).
		Return(dbProject, nil)

	for _, cacheKey := range setCacheKeys {
		mockRedis.EXPECT().
			SetData(ctx, cacheKey, dbProject, 24*time.Hour).
			Return(nil)
	}

	result, err := cachedRepo.FindSystem(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "system", result.ID)
	assert.True(t, result.IsSystem)
}

func TestCachedProjectRepository_Create(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:       "proj-123",
		Name:     "Test Project",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Create(ctx, project).
		Return(nil)

	err := cachedRepo.Create(ctx, project)

	assert.NoError(t, err)
}

func TestCachedProjectRepository_Update(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:       "proj-123",
		Name:     "Test Project",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, project).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, project)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedProjectRepository_Delete(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	project := &entities.Project{
		ID:        projectID,
		AccountID: "acc-123",
		IsSystem:  false,
	}

	mockDBRepo.EXPECT().
		Delete(ctx, project).
		Return(1, nil)

	affected, err := cachedRepo.Delete(ctx, project)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedProjectRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedProjects := []entities.Project{
		{ID: "proj-1", Name: "Project 1"},
		{ID: "proj-2", Name: "Project 2"},
	}

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetProjectsRequest"), "acc-123").
		Return(expectedProjects, nil)

	req := &models.GetProjectsRequest{}
	result, err := cachedRepo.FindAll(ctx, req, "acc-123")

	assert.NoError(t, err)
	assert.Equal(t, len(expectedProjects), len(result))
}

func TestCachedProjectRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockProjectRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewProjectCacheRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetProjectsRequest"), "acc-123").
		Return(5, nil)

	req := &models.GetProjectsRequest{}
	result, err := cachedRepo.Count(ctx, req, "acc-123")

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}
