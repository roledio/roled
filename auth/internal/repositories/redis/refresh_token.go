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

type refreshTokenRepository struct {
	repo  interfaces.RefreshTokenRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewRefreshTokenRepository(repo interfaces.RefreshTokenRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.RefreshTokenRepository {
	if redis == nil {
		return repo
	}
	return &refreshTokenRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *refreshTokenRepository) FindByClientIDAndRefreshTokenHash(ctx context.Context, clientID, refreshTokenHash string) (*entities.RefreshToken, error) {
	cacheKey := rediskeys.RefreshTokenByClientIDAndTokenHash(clientID, refreshTokenHash)
	var refreshToken entities.RefreshToken

	found, err := r.redis.GetData(ctx, cacheKey, &refreshToken)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get refresh token from redis cache, falling back to DB", "error", err, "client_id", clientID)
	} else if found {
		return &refreshToken, nil
	}

	refreshTokenPtr, err := r.repo.FindByClientIDAndRefreshTokenHash(ctx, clientID, refreshTokenHash)
	if err != nil {
		return nil, err
	}
	if refreshTokenPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, refreshTokenPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache refresh token in redis", "error", setErr, "client_id", clientID)
	}

	return refreshTokenPtr, nil
}

func (r *refreshTokenRepository) Create(ctx context.Context, refreshToken *entities.RefreshToken) error {
	// Don't cache newly created data to avoid stale cache if transaction rolls back
	return r.repo.Create(ctx, refreshToken)
}

func (r *refreshTokenRepository) UpdateUsedRefreshToken(ctx context.Context, refreshToken *entities.RefreshToken) (int, error) {
	return r.repo.UpdateUsedRefreshToken(ctx, refreshToken)
}

func (r *refreshTokenRepository) UpdateAsRevoked(ctx context.Context, refreshToken *entities.RefreshToken) (int, error) {
	return r.repo.UpdateAsRevoked(ctx, refreshToken)
}
