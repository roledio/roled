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

func TestCachedAuthCodeRepository_FindByClientIDAndCodeHash_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAuthCodeRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	codeHash := "codehash-123"
	cacheKey := rediskeys.AuthCodeByClientIDAndCodeHash(clientID, codeHash)
	expectedAuthCode := &entities.AuthCode{
		ID:        "auth-123",
		ClientID:  clientID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.AuthCode")).
		Run(func(ctx context.Context, key string, dest any) {
			ac := dest.(*entities.AuthCode)
			*ac = *expectedAuthCode
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByClientIDAndCodeHash(ctx, clientID, codeHash)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAuthCode.ID, result.ID)
}

func TestCachedAuthCodeRepository_FindByClientIDAndCodeHash_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbAuthCode := &entities.AuthCode{
		ID:        "auth-123",
		ClientID:  "client-123",
		CodeHash:  "codehash-123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockDBRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAuthCodeRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	codeHash := "codehash-123"
	cacheKey := rediskeys.AuthCodeByClientIDAndCodeHash(clientID, codeHash)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.AuthCode")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByClientIDAndCodeHash(ctx, clientID, codeHash).
		Return(dbAuthCode, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbAuthCode, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByClientIDAndCodeHash(ctx, clientID, codeHash)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbAuthCode.ID, result.ID)
}

func TestCachedAuthCodeRepository_Create(t *testing.T) {
	ctx := context.Background()
	authCode := &entities.AuthCode{
		ID:        "auth-123",
		ClientID:  "client-123",
		CodeHash:  "codehash-123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockDBRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAuthCodeRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, authCode).
		Return(nil)

	err := cachedRepo.Create(ctx, authCode)

	assert.NoError(t, err)
}

func TestCachedAuthCodeRepository_UpdateUsedAuthCode(t *testing.T) {
	ctx := context.Background()
	authCode := &entities.AuthCode{
		ID:        "auth-123",
		ClientID:  "client-123",
		CodeHash:  "codehash-123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockDBRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAuthCodeRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		UpdateUsedAuthCode(ctx, authCode).
		Return(1, nil)

	affected, err := cachedRepo.UpdateUsedAuthCode(ctx, authCode)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}
