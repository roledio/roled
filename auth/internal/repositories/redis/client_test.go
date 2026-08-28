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

func TestCachedClientRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	cacheKey := rediskeys.ClientByID(clientID)
	expectedClient := &entities.Client{
		ID:        clientID,
		Name:      "Test Client",
		ProjectID: "proj-123",
		IsDefault: false,
		IsActive:  true,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Client")).
		Run(func(ctx context.Context, key string, dest any) {
			cl := dest.(*entities.Client)
			*cl = *expectedClient
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, clientID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedClient.ID, result.ID)
}

func TestCachedClientRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbClient := &entities.Client{
		ID:        "client-123",
		Name:      "Test Client",
		ProjectID: "proj-123",
		IsDefault: false,
		IsActive:  true,
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	mainCacheKey := rediskeys.ClientByID(dbClient.ID)
	setCacheKeys := []string{
		mainCacheKey,
		rediskeys.ClientByIDAndProjectID(dbClient.ID, dbClient.ProjectID),
	}

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*entities.Client")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByID(ctx, dbClient.ID).
		Return(dbClient, nil)

	for _, cacheKey := range setCacheKeys {
		mockRedis.EXPECT().
			SetData(ctx, cacheKey, dbClient, 24*time.Hour).
			Return(nil)
	}

	result, err := cachedRepo.FindByID(ctx, clientID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbClient.ID, result.ID)
}

func TestCachedClientRepository_FindByProjectIDAndIsDefault_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	isDefault := true
	cacheKey := rediskeys.ClientByProjectIDAndIsDefault(projectID, isDefault)
	expectedClient := &entities.Client{
		ID:        "client-123",
		Name:      "Default Client",
		ProjectID: projectID,
		IsDefault: isDefault,
		IsActive:  true,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Client")).
		Run(func(ctx context.Context, key string, dest any) {
			cl := dest.(*entities.Client)
			*cl = *expectedClient
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndIsDefault(ctx, projectID, isDefault)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedClient.ID, result.ID)
}

func TestCachedClientRepository_FindByProjectIDAndIsDefault_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbClient := &entities.Client{
		ID:        "client-123",
		Name:      "Default Client",
		ProjectID: "proj-123",
		IsDefault: true,
		IsActive:  true,
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	isDefault := true
	mainCacheKey := rediskeys.ClientByProjectIDAndIsDefault(projectID, isDefault)
	setCacheKeys := []string{
		mainCacheKey,
		rediskeys.ClientByID(dbClient.ID),
		rediskeys.ClientByIDAndProjectID(dbClient.ID, dbClient.ProjectID),
	}

	mockRedis.EXPECT().
		GetData(ctx, mainCacheKey, mock.AnythingOfType("*entities.Client")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByProjectIDAndIsDefault(ctx, projectID, isDefault).
		Return(dbClient, nil)

	for _, cacheKey := range setCacheKeys {
		mockRedis.EXPECT().
			SetData(ctx, cacheKey, dbClient, 24*time.Hour).
			Return(nil)
	}

	result, err := cachedRepo.FindByProjectIDAndIsDefault(ctx, projectID, isDefault)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbClient.ID, result.ID)
}

func TestCachedClientRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		Name:      "Test Client",
		ProjectID: "proj-123",
		IsDefault: false,
		IsActive:  true,
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Create(ctx, client).
		Return(nil)

	err := cachedRepo.Create(ctx, client)

	assert.NoError(t, err)
}

func TestCachedClientRepository_Update(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		Name:      "Updated Client",
		ProjectID: "proj-123",
		IsDefault: true,
		IsActive:  true,
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, client).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, client)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedClientRepository_Delete(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	client := &entities.Client{
		ID:        clientID,
		ProjectID: "proj-123",
		IsDefault: true,
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Delete(ctx, client).
		Return(1, nil)

	affected, err := cachedRepo.Delete(ctx, client)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedClientRepository_DeleteByProjectID(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByProjectID(ctx, projectID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedClientRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetClientsRequest")).
		Return(5, nil)

	req := &models.GetClientsRequest{}
	result, err := cachedRepo.Count(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCachedClientRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedClients := []entities.Client{
		{ID: "client-1", Name: "Client 1"},
		{ID: "client-2", Name: "Client 2"},
	}

	mockDBRepo := repositorymocks.NewMockClientRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewClientRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetClientsRequest")).
		Return(expectedClients, nil)

	req := &models.GetClientsRequest{}
	result, err := cachedRepo.FindAll(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedClients), len(result))
}
