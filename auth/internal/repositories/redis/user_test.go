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
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedUserRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	userID := "user-123"
	cacheKey := rediskeys.UserByID(userID)
	userEmail := "test@example.com"
	displayName := "Test User"
	externalUserID := "ext-123"
	expectedUser := &entities.User{
		ID:              userID,
		ProjectID:       "proj-123",
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.User")).
		Run(func(ctx context.Context, key string, dest any) {
			u := dest.(*entities.User)
			*u = *expectedUser
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
}

func TestCachedUserRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	userEmail := "test@example.com"
	displayName := "Test User"
	externalUserID := "ext-123"
	dbUser := &entities.User{
		ID:              "user-123",
		ProjectID:       "proj-123",
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	userID := "user-123"
	cacheKey := rediskeys.UserByID(userID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.User")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByID(ctx, userID).
		Return(dbUser, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbUser, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbUser.ID, result.ID)
}

func TestCachedUserRepository_FindByProjectIDAndEmail_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	userEmail := "test@example.com"
	cacheKey := rediskeys.UserByProjectIDAndEmail(projectID, userEmail)
	displayName := "Test User"
	externalUserID := "ext-123"
	expectedUser := &entities.User{
		ID:              "user-123",
		ProjectID:       projectID,
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.User")).
		Run(func(ctx context.Context, key string, dest any) {
			u := dest.(*entities.User)
			*u = *expectedUser
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndEmail(ctx, projectID, userEmail)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
}

func TestCachedUserRepository_FindByProjectIDAndEmail_CacheMiss(t *testing.T) {
	ctx := context.Background()
	userEmail := "test@example.com"
	displayName := "Test User"
	externalUserID := "ext-123"
	dbUser := &entities.User{
		ID:              "user-123",
		ProjectID:       "proj-123",
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	cacheKey := rediskeys.UserByProjectIDAndEmail(projectID, userEmail)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.User")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByProjectIDAndEmail(ctx, projectID, userEmail).
		Return(dbUser, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbUser, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByProjectIDAndEmail(ctx, projectID, userEmail)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbUser.ID, result.ID)
}

func TestCachedUserRepository_FindByProjectIDAndExternalUserID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	projectID := "proj-123"
	externalUserID := "ext-123"
	cacheKey := rediskeys.UserByProjectIDAndExternalUserID(projectID, externalUserID)
	userEmail := "test@example.com"
	displayName := "Test User"
	expectedUser := &entities.User{
		ID:              "user-123",
		ProjectID:       projectID,
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.User")).
		Run(func(ctx context.Context, key string, dest any) {
			u := dest.(*entities.User)
			*u = *expectedUser
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByProjectIDAndExternalUserID(ctx, projectID, externalUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.ID, result.ID)
}

func TestCachedUserRepository_SetEmailVerified(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		SetEmailVerified(ctx, userID).
		Return(1, nil)

	affected, err := cachedRepo.SetEmailVerified(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	passwordHash := "newhash123"

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		UpdatePassword(ctx, userID, passwordHash).
		Return(1, nil)

	affected, err := cachedRepo.UpdatePassword(ctx, userID, passwordHash)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	userEmail := "updated@example.com"
	displayName := "Updated User"
	externalUserID := "ext-123"
	user := &entities.User{
		ID:              "user-123",
		ProjectID:       "proj-123",
		Email:           &userEmail,
		DisplayName:     displayName,
		ExternalUserID:  &externalUserID,
		AvatarURL:       nil,
		IsActive:        true,
		EmailVerifiedAt: nil,
	}

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, user).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByID(ctx, userID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_DeleteByAccountID(t *testing.T) {
	ctx := context.Background()
	accountID := "acc-123"

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByAccountID
	mockDBRepo.EXPECT().
		DeleteByAccountID(ctx, accountID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByAccountID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_DeleteByProjectID(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByProjectID
	mockDBRepo.EXPECT().
		DeleteByProjectID(ctx, projectID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedUserRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Count (cache should not be used for count queries)
	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetUsersRequest")).
		Return(5, nil)

	req := &models.GetUsersRequest{}
	result, err := cachedRepo.Count(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCachedUserRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedUsers := []interfaces.UserAndRole{
		{User: entities.User{ID: "user-1", Email: nil}, RoleID: "role-1", RoleName: "Admin"},
		{User: entities.User{ID: "user-2", Email: nil}, RoleID: "role-2", RoleName: "User"},
	}

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindAll (cache should not be used for paginated lists)
	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetUsersRequest")).
		Return(expectedUsers, nil)

	req := &models.GetUsersRequest{}
	result, err := cachedRepo.FindAll(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedUsers), len(result))
}

func TestCachedUserRepository_FindByProjectIDAndExternalUserIDJoinRole_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindByProjectIDAndExternalUserIDJoinRole (cache should not be used for join queries)
	mockDBRepo.EXPECT().
		FindByProjectIDAndExternalUserIDJoinRole(ctx, "proj-123", "ext-123").
		Return(&interfaces.UserAndRole{}, nil)

	result, err := cachedRepo.FindByProjectIDAndExternalUserIDJoinRole(ctx, "proj-123", "ext-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCachedUserRepository_FindByIDAndProjectIDJoinRole_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewUserRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindByIDAndProjectIDJoinRole (cache should not be used for join queries)
	mockDBRepo.EXPECT().
		FindByIDAndProjectIDJoinRole(ctx, "user-123", "proj-123").
		Return(&interfaces.UserAndRole{}, nil)

	result, err := cachedRepo.FindByIDAndProjectIDJoinRole(ctx, "user-123", "proj-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
}
