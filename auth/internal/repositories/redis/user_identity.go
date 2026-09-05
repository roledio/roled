package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/services/infra"
)

type userIdentityRepository struct {
	repo  interfaces.UserIdentityRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewUserIdentityRepository(repo interfaces.UserIdentityRepository, redis infra.RedisService, ttl time.Duration) interfaces.UserIdentityRepository {
	if redis == nil {
		return repo
	}
	return &userIdentityRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *userIdentityRepository) Create(ctx context.Context, userIdentity *entities.UserIdentity) error {
	return r.repo.Create(ctx, userIdentity)
}

func (r *userIdentityRepository) FindByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*entities.UserIdentity, error) {
	cacheKey := rediskeys.UserIdentityByProviderAndProviderUserID(provider, providerUserID)
	var userIdentity entities.UserIdentity

	found, err := r.redis.GetData(ctx, cacheKey, &userIdentity)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get user identity from redis cache, falling back to DB", "error", err, "provider", provider)
	} else if found {
		return &userIdentity, nil
	}

	userIdentityPtr, err := r.repo.FindByProviderAndProviderUserID(ctx, provider, providerUserID)
	if err != nil {
		return nil, err
	}
	if userIdentityPtr == nil {
		return nil, nil
	}

	// Cache with all possible keys
	cacheKeys := []string{
		cacheKey,
		rediskeys.UserIdentityByID(userIdentityPtr.ID),
	}

	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, userIdentityPtr, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache user identity in redis", "error", setErr, "provider", provider)
		}
	}

	return userIdentityPtr, nil
}

func (r *userIdentityRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.UserIdentity, error) {
	return r.repo.FindByUserID(ctx, userID)
}

func (r *userIdentityRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	return r.repo.DeleteByID(ctx, id)
}
