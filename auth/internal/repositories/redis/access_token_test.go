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
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedAccessTokenRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	tokenID := "token-123"
	cacheKey := rediskeys.AccessTokenByID(tokenID)
	expiresIn := 3600
	issuedAt := time.Now()
	expectedToken := &entities.AccessToken{
		ID:        tokenID,
		ProjectID: "proj-123",
		ClientID:  "client-123",
		GrantType: "authorization_code",
		Status:    constants.AccessTokenStatusIssued,
		ExpiresIn: &expiresIn,
		IssuedAt:  &issuedAt,
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.AccessToken")).
		Run(func(ctx context.Context, key string, dest any) {
			at := dest.(*entities.AccessToken)
			*at = *expectedToken
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, tokenID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedToken.ID, result.ID)
}

func TestCachedAccessTokenRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	expiresIn := 3600
	issuedAt := time.Now()
	dbToken := &entities.AccessToken{
		ID:        "token-123",
		ProjectID: "proj-123",
		ClientID:  "client-123",
		GrantType: "authorization_code",
		Status:    constants.AccessTokenStatusIssued,
		ExpiresIn: &expiresIn,
		IssuedAt:  &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	tokenID := "token-123"
	cacheKey := rediskeys.AccessTokenByID(tokenID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.AccessToken")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByID(ctx, tokenID).
		Return(dbToken, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbToken, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByID(ctx, tokenID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbToken.ID, result.ID)
}

func TestCachedAccessTokenRepository_Create(t *testing.T) {
	ctx := context.Background()
	expiresIn := 3600
	issuedAt := time.Now()
	token := &entities.AccessToken{
		ID:        "token-123",
		ProjectID: "proj-123",
		ClientID:  "client-123",
		GrantType: "authorization_code",
		Status:    constants.AccessTokenStatusIssued,
		ExpiresIn: &expiresIn,
		IssuedAt:  &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB Create
	mockDBRepo.EXPECT().
		Create(ctx, token).
		Return(nil)

	err := cachedRepo.Create(ctx, token)

	assert.NoError(t, err)
}

func TestCachedAccessTokenRepository_UpdateAsIssued(t *testing.T) {
	ctx := context.Background()
	expiresIn := 3600
	issuedAt := time.Now()
	token := &entities.AccessToken{
		ID:        "token-123",
		ProjectID: "proj-123",
		ClientID:  "client-123",
		GrantType: "authorization_code",
		Status:    constants.AccessTokenStatusIssued,
		ExpiresIn: &expiresIn,
		IssuedAt:  &issuedAt,
	}

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		UpdateAsIssued(ctx, token).
		Return(1, nil)

	affected, err := cachedRepo.UpdateAsIssued(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_UpdateAsRevoked(t *testing.T) {
	ctx := context.Background()
	tokenID := "token-123"

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		UpdateAsRevoked(ctx, tokenID).
		Return(1, nil)

	affected, err := cachedRepo.UpdateAsRevoked(ctx, tokenID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_DeleteByAccountID(t *testing.T) {
	ctx := context.Background()
	accountID := "acc-123"

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByAccountID
	mockDBRepo.EXPECT().
		DeleteByAccountID(ctx, accountID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByAccountID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_DeleteByUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByUserID
	mockDBRepo.EXPECT().
		DeleteByUserID(ctx, userID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_DeleteByProjectID(t *testing.T) {
	ctx := context.Background()
	projectID := "proj-123"

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByProjectID
	mockDBRepo.EXPECT().
		DeleteByProjectID(ctx, projectID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByProjectID(ctx, projectID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_DeleteByClientID(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	// Mock DB DeleteByClientID
	mockDBRepo.EXPECT().
		DeleteByClientID(ctx, clientID).
		Return(1, nil)

	affected, err := cachedRepo.DeleteByClientID(ctx, clientID)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedAccessTokenRepository_FindByIDJoin_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	tokenID := "token-123"
	cacheKey := rediskeys.AccessTokenByIDJoin(tokenID)
	expiresIn := 3600
	issuedAt := time.Now()
	expectedToken := &interfaces.AccessTokenJoinResult{
		ID:             tokenID,
		IssuedAt:       issuedAt,
		ExpiresIn:      expiresIn,
		ProjectID:      "proj-123",
		ProjectName:    "Project 123",
		ProjectLogoURL: nil,
		ClientID:       "client-123",
		ClientName:     "Client 123",
	}

	// Mock Redis GetData returning cache hit
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*interfaces.AccessTokenJoinResult")).
		Run(func(ctx context.Context, key string, dest any) {
			at := dest.(*interfaces.AccessTokenJoinResult)
			*at = *expectedToken
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByIDJoin(ctx, tokenID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedToken.ID, result.ID)
}

func TestCachedAccessTokenRepository_FindByIDJoin_CacheMiss(t *testing.T) {
	ctx := context.Background()
	expiresIn := 3600
	issuedAt := time.Now()
	dbToken := &interfaces.AccessTokenJoinResult{
		ID:             "token-123",
		IssuedAt:       issuedAt,
		ExpiresIn:      expiresIn,
		ProjectID:      "proj-123",
		ProjectName:    "Project 123",
		ProjectLogoURL: nil,
		ClientID:       "client-123",
		ClientName:     "Client 123",
	}

	mockDBRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewAccessTokenRepository(mockDBRepo, mockRedis, 24*time.Hour)

	tokenID := "token-123"
	cacheKey := rediskeys.AccessTokenByIDJoin(tokenID)

	// Mock Redis GetData returning cache miss
	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*interfaces.AccessTokenJoinResult")).
		Return(false, nil)

	// Mock DB query
	mockDBRepo.EXPECT().
		FindByIDJoin(ctx, tokenID).
		Return(dbToken, nil)

	// Mock Redis SetData saving to cache
	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbToken, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByIDJoin(ctx, tokenID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbToken.ID, result.ID)
}
