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

type oAuthConnectionRepository struct {
	repo  interfaces.OAuthConnectionRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewOAuthConnectionRepository(repo interfaces.OAuthConnectionRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.OAuthConnectionRepository {
	if redis == nil {
		return repo
	}
	return &oAuthConnectionRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *oAuthConnectionRepository) FindByProjectID(ctx context.Context, projectID string) ([]entities.OAuthConnection, error) {
	cacheKey := rediskeys.OAuthConnectionsByProjectID(projectID)
	var connections []entities.OAuthConnection

	found, err := r.redis.GetData(ctx, cacheKey, &connections)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get project OAuth connections from redis cache, falling back to DB", "error", err, "project_id", projectID)
	} else if found {
		return connections, nil
	}

	connections, err = r.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if setErr := r.redis.SetData(ctx, cacheKey, connections, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache project OAuth connections in redis",
			"error", setErr,
			"project_id", projectID,
			"cache_key", cacheKey)
	}

	for _, conn := range connections {
		individualCacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, conn.Provider)
		if setErr := r.redis.SetData(ctx, individualCacheKey, conn, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache individual project OAuth connection in redis",
				"error", setErr,
				"project_id", projectID,
				"provider", conn.Provider,
				"cache_key", individualCacheKey)
		}
	}

	return connections, nil
}

func (r *oAuthConnectionRepository) FindByProjectIDAndProvider(ctx context.Context, projectID, provider string) (*entities.OAuthConnection, error) {
	cacheKey := rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider)

	var conn entities.OAuthConnection
	found, err := r.redis.GetData(ctx, cacheKey, &conn)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get OAuth connection from redis cache, falling back to DB", "error", err, "project_id", projectID, "provider", provider)
	} else if found {
		return &conn, nil
	}

	connPtr, err := r.repo.FindByProjectIDAndProvider(ctx, projectID, provider)
	if err != nil {
		return nil, err
	}
	if connPtr == nil {
		return nil, nil
	}

	cacheKeys := []string{
		cacheKey,
		rediskeys.OAuthConnectionsByProjectID(projectID),
	}
	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, connPtr, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache OAuth connection in redis",
				"error", setErr,
				"project_id", projectID,
				"provider", provider,
				"cache_key", key)
		}
	}

	return connPtr, nil
}

func (r *oAuthConnectionRepository) Create(ctx context.Context, connection *entities.OAuthConnection) error {
	return r.repo.Create(ctx, connection)
}

func (r *oAuthConnectionRepository) Update(ctx context.Context, connection *entities.OAuthConnection) (int, error) {
	return r.repo.Update(ctx, connection)
}

func (r *oAuthConnectionRepository) Delete(ctx context.Context, projectID, provider string) (int, error) {
	return r.repo.Delete(ctx, projectID, provider)
}
