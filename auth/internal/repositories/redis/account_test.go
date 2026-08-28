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

func TestCachedAccountRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	accountID := "acc-123"
	cacheKey := rediskeys.AccountByID(accountID)
	expectedAccount := &entities.Account{
		ID:       "acc-123",
		Name:     "Test Account",
		IsSystem: false,
		IsActive: true,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Account")).
		Run(func(ctx context.Context, key string, dest any) {
			ac := dest.(*entities.Account)
			*ac = *expectedAccount
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, accountID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAccount.ID, result.ID)
	assert.Equal(t, expectedAccount.Name, result.Name)
}

func TestCachedAccountRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbAccount := &entities.Account{
		ID:       "acc-123",
		Name:     "Test Account",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	accountID := "acc-123"
	cacheKey := rediskeys.AccountByID(accountID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Account")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByID(ctx, accountID).
		Return(dbAccount, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbAccount, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByID(ctx, accountID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbAccount.ID, result.ID)
}

func TestCachedAccountRepository_Create(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "acc-123",
		Name:     "Test Account",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, account).
		Return(nil)

	err := cachedRepo.Create(ctx, account)

	assert.NoError(t, err)
}

func TestCachedAccountRepository_Update(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "acc-123",
		Name:     "Test Account",
		IsSystem: false,
		IsActive: true,
	}

	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Update(ctx, account).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccountRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	accountID := "acc-123"

	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByID(ctx, accountID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccountRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedAccounts := []entities.Account{
		{ID: "acc-1", Name: "Account 1"},
		{ID: "acc-2", Name: "Account 2"},
	}

	mockDBRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccountRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB FindAll (cache should not be used for paginated lists)
	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.Anything).
		Return(expectedAccounts, nil)

	req := &models.GetAccountsRequest{}
	result, err := cachedRepo.FindAll(ctx, req, nil)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedAccounts), len(result))
}
