package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedRefreshTokenRepository_FindByClientIDAndRefreshTokenHash_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRefreshTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	refreshTokenHash := "refreshtokenhash-123"
	cacheKey := rediskeys.RefreshTokenByClientIDAndTokenHash(clientID, refreshTokenHash)
	expiresIn := 86400
	issuedAt := time.Now()
	expectedToken := &entities.RefreshToken{
		ID:               "token-123",
		ClientID:         clientID,
		RefreshTokenHash: refreshTokenHash,
		Status:           constants.RefreshTokenStatusUsed,
		ExpiresIn:        &expiresIn,
		IssuedAt:         &issuedAt,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.RefreshToken")).
		Run(func(ctx context.Context, key string, dest any) {
			rt := dest.(*entities.RefreshToken)
			*rt = *expectedToken
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByClientIDAndRefreshTokenHash(ctx, clientID, refreshTokenHash)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedToken.ID, result.ID)
}

func TestCachedRefreshTokenRepository_FindByClientIDAndRefreshTokenHash_CacheMiss(t *testing.T) {
	ctx := context.Background()
	expiresIn := 86400
	issuedAt := time.Now()
	dbToken := &entities.RefreshToken{
		ID:               "token-123",
		ClientID:         "client-123",
		RefreshTokenHash: "refreshtokenhash-123",
		Status:           constants.RefreshTokenStatusUsed,
		ExpiresIn:        &expiresIn,
		IssuedAt:         &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRefreshTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	clientID := "client-123"
	refreshTokenHash := "refreshtokenhash-123"
	cacheKey := rediskeys.RefreshTokenByClientIDAndTokenHash(clientID, refreshTokenHash)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.RefreshToken")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByClientIDAndRefreshTokenHash(ctx, clientID, refreshTokenHash).
		Return(dbToken, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbToken, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByClientIDAndRefreshTokenHash(ctx, clientID, refreshTokenHash)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbToken.ID, result.ID)
}

func TestCachedRefreshTokenRepository_Create(t *testing.T) {
	ctx := context.Background()
	expiresIn := 86400
	issuedAt := time.Now()
	token := &entities.RefreshToken{
		ID:               "token-123",
		ClientID:         "client-123",
		RefreshTokenHash: "refreshtokenhash-123",
		Status:           constants.RefreshTokenStatusUsed,
		ExpiresIn:        &expiresIn,
		IssuedAt:         &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRefreshTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, token).
		Return(nil)

	err := cachedRepo.Create(ctx, token)

	assert.NoError(t, err)
}

func TestCachedRefreshTokenRepository_UpdateUsedRefreshToken(t *testing.T) {
	ctx := context.Background()
	expiresIn := 86400
	issuedAt := time.Now()
	token := &entities.RefreshToken{
		ID:               "token-123",
		ClientID:         "client-123",
		RefreshTokenHash: "refreshtokenhash-123",
		Status:           constants.RefreshTokenStatusUsed,
		ExpiresIn:        &expiresIn,
		IssuedAt:         &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRefreshTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		UpdateUsedRefreshToken(ctx, token).
		Return(1, nil)

	affected, err := cachedRepo.UpdateUsedRefreshToken(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedRefreshTokenRepository_UpdateAsRevoked(t *testing.T) {
	ctx := context.Background()
	tokenID := "token-123"
	clientID := "client-123"
	refreshTokenHash := "refreshtokenhash-123"

	mockDBRepo := repositorymocks.NewMockRefreshTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewRefreshTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockRefreshToken := &entities.RefreshToken{ID: tokenID, ClientID: clientID, RefreshTokenHash: refreshTokenHash}

	mockDBRepo.EXPECT().
		UpdateAsRevoked(ctx, mockRefreshToken).
		Return(1, nil)

	affected, err := cachedRepo.UpdateAsRevoked(ctx, mockRefreshToken)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}
