package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/roledio/roled/internal/services/infra"
)

type resourceRepository struct {
	repo  interfaces.ResourceRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewResourceRepository(repo interfaces.ResourceRepository, redis infra.RedisService, ttl time.Duration) interfaces.ResourceRepository {
	if redis == nil {
		return repo
	}
	return &resourceRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *resourceRepository) FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Resource, error) {
	cacheKey := rediskeys.ResourceByProjectIDAndCode(projectID, code)
	var resource entities.Resource

	found, err := r.redis.GetData(ctx, cacheKey, &resource)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get resource from redis cache, falling back to DB", "error", err, "project_id", projectID, "code", code)
	} else if found {
		return &resource, nil
	}

	resourcePtr, err := r.repo.FindByProjectIDAndCode(ctx, projectID, code)
	if err != nil {
		return nil, err
	}
	if resourcePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, resourcePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache resource in redis", "error", setErr, "project_id", projectID, "code", code)
	}

	return resourcePtr, nil
}

func (r *resourceRepository) Create(ctx context.Context, resources []entities.Resource) (int, error) {
	// Don't cache newly created data to avoid stale cache if transaction rolls back
	return r.repo.Create(ctx, resources)
}

func (r *resourceRepository) FindByIDAndProjectID(ctx context.Context, resourceID string, projectID string) (*entities.Resource, error) {
	cacheKey := rediskeys.ResourceByIDAndProjectID(resourceID, projectID)
	var resource entities.Resource

	found, err := r.redis.GetData(ctx, cacheKey, &resource)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get resource from redis cache, falling back to DB", "error", err, "project_id", projectID, "resource_id", resourceID)
	} else if found {
		return &resource, nil
	}

	resourcePtr, err := r.repo.FindByIDAndProjectID(ctx, resourceID, projectID)
	if err != nil {
		return nil, err
	}
	if resourcePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, resourcePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache resource in redis", "error", setErr, "resource_id", resourceID)
	}

	return resourcePtr, nil
}

func (r *resourceRepository) Update(ctx context.Context, resource *entities.Resource) (int, error) {
	return r.repo.Update(ctx, resource)
}

func (r *resourceRepository) Delete(ctx context.Context, resource *entities.Resource) (int, error) {
	return r.repo.Delete(ctx, resource)
}

func (r *resourceRepository) Count(ctx context.Context, req *models.GetResourcesRequest) (int, error) {
	// Don't cache count queries
	return r.repo.Count(ctx, req)
}

func (r *resourceRepository) FindAll(ctx context.Context, req *models.GetResourcesRequest) ([]entities.Resource, error) {
	// Don't cache paginated lists
	return r.repo.FindAll(ctx, req)
}
