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

type authCodeRepository struct {
	repo  interfaces.AuthCodeRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewAuthCodeRepository(repo interfaces.AuthCodeRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.AuthCodeRepository {
	if redis == nil {
		return repo
	}
	return &authCodeRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *authCodeRepository) FindByClientIDAndCodeHash(ctx context.Context, clientID string, codeHash string) (*entities.AuthCode, error) {
	cacheKey := rediskeys.AuthCodeByClientIDAndCodeHash(clientID, codeHash)
	var authCode entities.AuthCode

	found, err := r.redis.GetData(ctx, cacheKey, &authCode)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get auth code from redis cache, falling back to DB", "error", err, "client_id", clientID)
	} else if found {
		return &authCode, nil
	}

	authCodePtr, err := r.repo.FindByClientIDAndCodeHash(ctx, clientID, codeHash)
	if err != nil {
		return nil, err
	}
	if authCodePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, authCodePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache auth code in redis", "error", setErr, "client_id", clientID)
	}

	return authCodePtr, nil
}

func (r *authCodeRepository) Create(ctx context.Context, authCode *entities.AuthCode) error {
	return r.repo.Create(ctx, authCode)
}

func (r *authCodeRepository) UpdateUsedAuthCode(ctx context.Context, authCode *entities.AuthCode) (int, error) {
	return r.repo.UpdateUsedAuthCode(ctx, authCode)
}
