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
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories/redis"
)

func TestCachedResourceRepository_FindByProjectIDAndCode_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	code := "users"
	cacheKey := rediskeys.ResourceByProjectIDAndCode(projectID, code)
	description := "User Management"
	expectedResource := &entities.Resource{
		ID:          "res-1",
		ProjectID:   projectID,
		Code:        code,
		Name:        "Users",
		Description: &description,
		IsDefault:   false,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Resource")).
		Run(func(ctx context.Context, key string, dest any) {
			r := dest.(*entities.Resource)
			*r = *expectedResource
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndCode(ctx, projectID, code)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResource.ID, result.ID)
}

func TestCachedResourceRepository_FindByProjectIDAndCode_CacheMiss(t *testing.T) {
	ctx := context.Background()
	description := "User Management"
	dbResource := &entities.Resource{
		ID:          "res-1",
		ProjectID:   "proj-123",
		Code:        "users",
		Name:        "Users",
		Description: &description,
		IsDefault:   false,
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	code := "users"
	cacheKey := rediskeys.ResourceByProjectIDAndCode(projectID, code)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Resource")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectIDAndCode(ctx, projectID, code).
		Return(dbResource, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbResource, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndCode(ctx, projectID, code)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbResource.ID, result.ID)
}

func TestCachedResourceRepository_Create(t *testing.T) {
	ctx := context.Background()
	description := "User Management"
	resources := []entities.Resource{
		{ID: "res-1", ProjectID: "proj-123", Code: "users", Name: "Users", Description: &description},
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, resources).
		Return(1, nil)

	affected, err := cachedRepo.Create(ctx, resources)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedResourceRepository_FindByIDAndProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	resourceID := "res-1"
	projectID := "proj-123"
	cacheKey := rediskeys.ResourceByIDAndProjectID(resourceID, projectID)
	description := "User Management"
	expectedResource := &entities.Resource{
		ID:          resourceID,
		ProjectID:   projectID,
		Code:        "users",
		Name:        "Users",
		Description: &description,
		IsDefault:   false,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Resource")).
		Run(func(ctx context.Context, key string, dest any) {
			r := dest.(*entities.Resource)
			*r = *expectedResource
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByIDAndProjectID(ctx, resourceID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedResource.ID, result.ID)
}

func TestCachedResourceRepository_FindByIDAndProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	description := "User Management"
	dbResource := &entities.Resource{
		ID:          "res-1",
		ProjectID:   "proj-123",
		Code:        "users",
		Name:        "Users",
		Description: &description,
		IsDefault:   false,
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	resourceID := "res-1"
	projectID := "proj-123"
	cacheKey := rediskeys.ResourceByIDAndProjectID(resourceID, projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Resource")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByIDAndProjectID(ctx, resourceID, projectID).
		Return(dbResource, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbResource, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByIDAndProjectID(ctx, resourceID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbResource.ID, result.ID)
}

func TestCachedResourceRepository_Update(t *testing.T) {
	ctx := context.Background()
	description := "Role Management"
	resource := &entities.Resource{
		ID:          "res-1",
		ProjectID:   "proj-123",
		Code:        "roles",
		Name:        "Roles",
		Description: &description,
		IsDefault:   false,
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, resource).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, resource)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedResourceRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	resourceID := "res-1"
	description := "User Management"
	resource := &entities.Resource{
		ID:          resourceID,
		ProjectID:   "proj-123",
		Code:        "users",
		Description: &description,
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Delete(ctx, resource).
		Return(1, nil)

	affected, err := cachedRepo.Delete(ctx, resource)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedResourceRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Count (cache should not be used for count queries)
	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetResourcesRequest")).
		Return(5, nil)

	req := &models.GetResourcesRequest{}
	result, err := cachedRepo.Count(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCachedResourceRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedResources := []entities.Resource{
		{ID: "res-1", ProjectID: "proj-123", Code: "users", Name: "Users"},
		{ID: "res-2", ProjectID: "proj-123", Code: "roles", Name: "Roles"},
	}

	mockDBRepo := repositorymocks.NewMockResourceRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewResourceRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindAll (cache should not be used for paginated lists)
	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetResourcesRequest")).
		Return(expectedResources, nil)

	req := &models.GetResourcesRequest{}
	result, err := cachedRepo.FindAll(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedResources), len(result))
}
