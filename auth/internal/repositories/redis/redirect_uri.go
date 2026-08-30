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

type redirectURIRepository struct {
	repo  interfaces.RedirectURIRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewRedirectURIRepository(repo interfaces.RedirectURIRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.RedirectURIRepository {
	if redis == nil {
		return repo
	}
	return &redirectURIRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *redirectURIRepository) FindByProjectIDAndRedirectURI(ctx context.Context, projectID string, redirectURI string) (*entities.RedirectURI, error) {
	cacheKey := rediskeys.RedirectURIByProjectIDAndURI(projectID, redirectURI)
	var redirectURIEntity entities.RedirectURI

	found, err := r.redis.GetData(ctx, cacheKey, &redirectURIEntity)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get redirect URI from redis cache, falling back to DB", "error", err, "project_id", projectID)
	} else if found {
		return &redirectURIEntity, nil
	}

	redirectURIPtr, err := r.repo.FindByProjectIDAndRedirectURI(ctx, projectID, redirectURI)
	if err != nil {
		return nil, err
	}
	if redirectURIPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, redirectURIPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache redirect URI in redis", "error", setErr, "project_id", projectID)
	}

	return redirectURIPtr, nil
}

func (r *redirectURIRepository) FindByProjectID(ctx context.Context, projectID string) ([]entities.RedirectURI, error) {
	cacheKey := rediskeys.RedirectURIsByProjectID(projectID)
	var redirectURIs []entities.RedirectURI

	found, err := r.redis.GetData(ctx, cacheKey, &redirectURIs)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get redirect URIs from redis cache, falling back to DB", "error", err, "project_id", projectID)
	} else if found {
		return redirectURIs, nil
	}

	redirectURIs, err = r.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if len(redirectURIs) > 0 {
		if setErr := r.redis.SetData(ctx, cacheKey, redirectURIs, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache redirect URIs in redis", "error", setErr, "project_id", projectID)
		}
		// Cache each redirect URI individually
		for _, redirectURI := range redirectURIs {
			cacheKey := rediskeys.RedirectURIByProjectIDAndURI(projectID, redirectURI.RedirectURI)
			if setErr := r.redis.SetData(ctx, cacheKey, redirectURI, r.ttl); setErr != nil {
				log.WithContext(ctx).Warnw("Failed to cache redirect URI in redis", "error", setErr, "project_id", projectID)
			}
		}
	}

	return redirectURIs, nil
}

func (r *redirectURIRepository) Create(ctx context.Context, redirectURIs []entities.RedirectURI) error {
	return r.repo.Create(ctx, redirectURIs)
}

func (r *redirectURIRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	return r.repo.DeleteByProjectID(ctx, projectID)
}
