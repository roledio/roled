package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/redis"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
)

func TestGoogleOAuthTransactionRepository_Store_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	transaction := &models.GoogleOAuthTransaction{
		ClientID:            "client-123",
		RedirectURI:         "https://example.com/callback",
		Scope:               "openid profile email",
		State:               "state-abc123",
		CodeChallenge:       "challenge-xyz",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now().UTC(),
	}

	cacheKey := rediskeys.GoogleOAuthTransaction(transaction.State)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, transaction, 5*time.Minute).
		Return(nil)

	err := repo.Store(ctx, transaction)

	assert.NoError(t, err)
}

func TestGoogleOAuthTransactionRepository_Store_RedisError(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	transaction := &models.GoogleOAuthTransaction{
		ClientID:            "client-123",
		RedirectURI:         "https://example.com/callback",
		Scope:               "openid profile email",
		State:               "state-abc123",
		CodeChallenge:       "challenge-xyz",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now().UTC(),
	}

	cacheKey := rediskeys.GoogleOAuthTransaction(transaction.State)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, transaction, 5*time.Minute).
		Return(assert.AnError)

	err := repo.Store(ctx, transaction)

	assert.Error(t, err)
}

func TestGoogleOAuthTransactionRepository_Retrieve_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	expectedTransaction := &models.GoogleOAuthTransaction{
		ClientID:            "client-123",
		RedirectURI:         "https://example.com/callback",
		Scope:               "openid profile email",
		State:               state,
		CodeChallenge:       "challenge-xyz",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now().UTC(),
	}

	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*models.GoogleOAuthTransaction")).
		Run(func(ctx context.Context, key string, dest any) {
			tx := dest.(*models.GoogleOAuthTransaction)
			*tx = *expectedTransaction
		}).
		Return(true, nil)

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{cacheKey}).
		Return(nil)

	result, err := repo.Retrieve(ctx, state)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTransaction.ClientID, result.ClientID)
	assert.Equal(t, expectedTransaction.State, result.State)
}

func TestGoogleOAuthTransactionRepository_Retrieve_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*models.GoogleOAuthTransaction")).
		Return(false, nil)

	result, err := repo.Retrieve(ctx, state)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGoogleOAuthTransactionRepository_Retrieve_RedisError(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*models.GoogleOAuthTransaction")).
		Return(false, assert.AnError)

	result, err := repo.Retrieve(ctx, state)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGoogleOAuthTransactionRepository_Retrieve_DeleteError(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	expectedTransaction := &models.GoogleOAuthTransaction{
		ClientID:            "client-123",
		RedirectURI:         "https://example.com/callback",
		Scope:               "openid profile email",
		State:               state,
		CodeChallenge:       "challenge-xyz",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now().UTC(),
	}

	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*models.GoogleOAuthTransaction")).
		Run(func(ctx context.Context, key string, dest any) {
			tx := dest.(*models.GoogleOAuthTransaction)
			*tx = *expectedTransaction
		}).
		Return(true, nil)

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{cacheKey}).
		Return(assert.AnError)

	result, err := repo.Retrieve(ctx, state)

	// Should not return error even if deletion fails (as per implementation comment)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTransaction.ClientID, result.ClientID)
}

func TestGoogleOAuthTransactionRepository_Delete_Success(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{cacheKey}).
		Return(nil)

	err := repo.Delete(ctx, state)

	assert.NoError(t, err)
}

func TestGoogleOAuthTransactionRepository_Delete_Error(t *testing.T) {
	ctx := context.Background()
	mockRedis := redismocks.NewMockService(t)

	repo := redis.NewGoogleOAuthTransactionRepository(mockRedis, 5*time.Minute)

	state := "state-abc123"
	cacheKey := rediskeys.GoogleOAuthTransaction(state)

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{cacheKey}).
		Return(assert.AnError)

	err := repo.Delete(ctx, state)

	// Should not return error even if deletion fails (as per implementation comment)
	assert.NoError(t, err)
}
