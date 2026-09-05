package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/infra"
)

// GoogleOAuthTransactionRepository handles storing and retrieving Google OAuth transactions from Redis
type GoogleOAuthTransactionRepository struct {
	redis infra.RedisService
	ttl   time.Duration
}

func NewGoogleOAuthTransactionRepository(redis infra.RedisService, ttl time.Duration) *GoogleOAuthTransactionRepository {
	return &GoogleOAuthTransactionRepository{
		redis: redis,
		ttl:   ttl,
	}
}

// Store stores the Google OAuth transaction in Redis with a TTL
// Returns an error if storage fails
func (r *GoogleOAuthTransactionRepository) Store(ctx context.Context, transaction *models.GoogleOAuthTransaction) error {
	cacheKey := rediskeys.GoogleOAuthTransaction(transaction.State)
	if err := r.redis.SetData(ctx, cacheKey, transaction, r.ttl); err != nil {
		log.WithContext(ctx).Errorw("Failed to store Google OAuth transaction in Redis", "error", err, "state", transaction.State)
		return err
	}
	return nil
}

// Retrieve retrieves the Google OAuth transaction from Redis by state
// Returns nil if not found or an error if retrieval fails
// This operation is destructive - the transaction is deleted after retrieval to ensure single-use
func (r *GoogleOAuthTransactionRepository) Retrieve(ctx context.Context, state string) (*models.GoogleOAuthTransaction, error) {
	cacheKey := rediskeys.GoogleOAuthTransaction(state)
	var transaction models.GoogleOAuthTransaction

	found, err := r.redis.GetData(ctx, cacheKey, &transaction)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to retrieve Google OAuth transaction from Redis", "error", err, "state", state)
		return nil, err
	}

	if !found {
		return nil, nil
	}

	// Delete the transaction immediately to ensure it's single-use
	if err := r.redis.DeleteManyWithContext(ctx, []string{cacheKey}); err != nil {
		log.WithContext(ctx).Warnw("Failed to delete Google OAuth transaction from Redis after retrieval", "error", err, "state", state)
		// Don't return error here - the data was successfully retrieved, deletion failure is non-critical
	}

	return &transaction, nil
}

// Delete removes the Google OAuth transaction from Redis by state
// Used for cleanup in error scenarios
func (r *GoogleOAuthTransactionRepository) Delete(ctx context.Context, state string) error {
	cacheKey := rediskeys.GoogleOAuthTransaction(state)
	if err := r.redis.DeleteManyWithContext(ctx, []string{cacheKey}); err != nil {
		log.WithContext(ctx).Warnw("Failed to delete Google OAuth transaction from Redis", "error", err, "state", state)
		// Don't return error - this is best-effort cleanup
	}
	return nil
}
