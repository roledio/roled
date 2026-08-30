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

type accessTokenRepository struct {
	repo  interfaces.AccessTokenRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewAccessTokenRepository(repo interfaces.AccessTokenRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.AccessTokenRepository {
	if redis == nil {
		return repo
	}
	return &accessTokenRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *accessTokenRepository) FindByID(ctx context.Context, id string) (*entities.AccessToken, error) {
	cacheKey := rediskeys.AccessTokenByID(id)
	var token entities.AccessToken

	found, err := r.redis.GetData(ctx, cacheKey, &token)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get access token from redis cache, falling back to DB", "error", err, "token_id", id)
	} else if found {
		return &token, nil
	}

	tokenPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tokenPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, tokenPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache access token in redis", "error", setErr, "token_id", id)
	}

	return tokenPtr, nil
}

func (r *accessTokenRepository) Create(ctx context.Context, token *entities.AccessToken) error {
	return r.repo.Create(ctx, token)
}

func (r *accessTokenRepository) UpdateAsIssued(ctx context.Context, token *entities.AccessToken) (int, error) {
	return r.repo.UpdateAsIssued(ctx, token)
}

func (r *accessTokenRepository) UpdateAsRevoked(ctx context.Context, id string) (int, error) {
	return r.repo.UpdateAsRevoked(ctx, id)
}

func (r *accessTokenRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	return r.repo.DeleteByAccountID(ctx, accountID)
}

func (r *accessTokenRepository) DeleteByUserID(ctx context.Context, userID string) (int, error) {
	return r.repo.DeleteByUserID(ctx, userID)
}

func (r *accessTokenRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	return r.repo.DeleteByProjectID(ctx, projectID)
}

func (r *accessTokenRepository) DeleteByClientID(ctx context.Context, clientID string) (int, error) {
	return r.repo.DeleteByClientID(ctx, clientID)
}

func (r *accessTokenRepository) FindByIDJoin(ctx context.Context, id string) (*interfaces.AccessTokenJoinResult, error) {
	cacheKey := rediskeys.AccessTokenByIDJoin(id)
	var token interfaces.AccessTokenJoinResult

	found, err := r.redis.GetData(ctx, cacheKey, &token)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get access token (join) from redis cache, falling back to DB", "error", err, "token_id", id)
	} else if found {
		return &token, nil
	}

	tokenPtr, err := r.repo.FindByIDJoin(ctx, id)
	if err != nil {
		return nil, err
	}
	if tokenPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, tokenPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache access token (join) in redis", "error", setErr, "token_id", id)
	}

	return tokenPtr, nil
}
