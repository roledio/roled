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

func TestCachedRoleRepository_FindByProjectIDAndCode_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	code := "admin"
	cacheKey := rediskeys.RoleByProjectIDAndCode(projectID, code)
	expectedRole := &entities.Role{
		ID:          "role-123",
		ProjectID:   projectID,
		Code:        code,
		Name:        "Administrator",
		Description: "System Administrator",
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Role")).
		Run(func(ctx context.Context, key string, dest any) {
			r := dest.(*entities.Role)
			*r = *expectedRole
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndCode(ctx, projectID, code)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRole.ID, result.ID)
}

func TestCachedRoleRepository_FindByProjectIDAndCode_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbRole := &entities.Role{
		ID:          "role-123",
		ProjectID:   "proj-123",
		Code:        "admin",
		Name:        "Administrator",
		Description: "System Administrator",
	}

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	code := "admin"
	cacheKey := rediskeys.RoleByProjectIDAndCode(projectID, code)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Role")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectIDAndCode(ctx, projectID, code).
		Return(dbRole, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbRole, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndCode(ctx, projectID, code)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbRole.ID, result.ID)
}

func TestCachedRoleRepository_FindByIDAndProjectID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	roleID := "role-123"
	projectID := "proj-123"
	cacheKey := rediskeys.RoleByIDAndProjectID(roleID, projectID)
	expectedRole := &entities.Role{
		ID:          roleID,
		ProjectID:   projectID,
		Code:        "admin",
		Name:        "Administrator",
		Description: "System Administrator",
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Role")).
		Run(func(ctx context.Context, key string, dest any) {
			r := dest.(*entities.Role)
			*r = *expectedRole
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByIDAndProjectID(ctx, roleID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRole.ID, result.ID)
}

func TestCachedRoleRepository_FindByIDAndProjectID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbRole := &entities.Role{
		ID:          "role-123",
		ProjectID:   "proj-123",
		Code:        "admin",
		Name:        "Administrator",
		Description: "System Administrator",
	}

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	roleID := "role-123"
	projectID := "proj-123"
	cacheKey := rediskeys.RoleByIDAndProjectID(roleID, projectID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Role")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByIDAndProjectID(ctx, roleID, projectID).
		Return(dbRole, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbRole, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByIDAndProjectID(ctx, roleID, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbRole.ID, result.ID)
}

func TestCachedRoleRepository_Create(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:          "role-123",
		ProjectID:   "proj-123",
		Code:        "admin",
		Name:        "Administrator",
		Description: "System Administrator",
	}

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, role).
		Return(nil)

	err := cachedRepo.Create(ctx, role)

	assert.NoError(t, err)
}

func TestCachedRoleRepository_Update(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:          "role-123",
		ProjectID:   "proj-123",
		Code:        "user",
		Name:        "User",
		Description: "Regular User",
	}

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, role).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, role)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedRoleRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByID(ctx, roleID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByID(ctx, roleID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedRoleRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Count (cache should not be used for count queries)
	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetProjectRolesRequest")).
		Return(5, nil)

	req := &models.GetProjectRolesRequest{}
	result, err := cachedRepo.Count(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCachedRoleRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedRoles := []entities.Role{
		{ID: "role-1", ProjectID: "proj-123", Code: "admin", Name: "Admin"},
		{ID: "role-2", ProjectID: "proj-123", Code: "user", Name: "User"},
	}

	mockDBRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRoleRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindAll (cache should not be used for paginated lists)
	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetProjectRolesRequest")).
		Return(expectedRoles, nil)

	req := &models.GetProjectRolesRequest{}
	result, err := cachedRepo.FindAll(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedRoles), len(result))
}
