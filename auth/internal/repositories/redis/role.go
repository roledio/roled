package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/services/infra"
)

type roleRepository struct {
	repo  interfaces.RoleRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewRoleRepository(repo interfaces.RoleRepository, redis infra.RedisService, ttl time.Duration) interfaces.RoleRepository {
	if redis == nil {
		return repo
	}
	return &roleRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *roleRepository) FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Role, error) {
	cacheKey := rediskeys.RoleByProjectIDAndCode(projectID, code)
	var role entities.Role

	found, err := r.redis.GetData(ctx, cacheKey, &role)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get role from redis cache, falling back to DB",
			"error", err,
			"project_id", projectID,
			"code", code,
			"cache_key", cacheKey)
	} else if found {
		return &role, nil
	}

	rolePtr, err := r.repo.FindByProjectIDAndCode(ctx, projectID, code)
	if err != nil {
		return nil, err
	}
	if rolePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, rolePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache role in redis",
			"error", setErr,
			"project_id", projectID,
			"code", code,
			"cache_key", cacheKey,
		)
	}

	return rolePtr, nil
}

func (r *roleRepository) FindByUserID(ctx context.Context, userID string) (*entities.Role, error) {
	cacheKey := rediskeys.RoleByUserID(userID)
	var role entities.Role

	found, err := r.redis.GetData(ctx, cacheKey, &role)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get role from redis cache, falling back to DB",
			"error", err,
			"user_id", userID,
			"cache_key", cacheKey,
		)
	} else if found {
		return &role, nil
	}

	rolePtr, err := r.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if rolePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, rolePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache role in redis",
			"error", setErr,
			"user_id", userID,
			"cache_key", cacheKey,
		)
	}

	return rolePtr, nil
}

func (r *roleRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Role, error) {
	cacheKey := rediskeys.RoleByIDAndProjectID(id, projectID)
	var role entities.Role

	found, err := r.redis.GetData(ctx, cacheKey, &role)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get role from redis cache, falling back to DB",
			"error", err,
			"role_id", id,
			"project_id", projectID,
			"cache_key", cacheKey,
		)
	} else if found {
		return &role, nil
	}

	rolePtr, err := r.repo.FindByIDAndProjectID(ctx, id, projectID)
	if err != nil {
		return nil, err
	}
	if rolePtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, rolePtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache role in redis",
			"error", setErr,
			"role_id", id,
			"project_id", projectID,
			"cache_key", cacheKey,
		)
	}

	return rolePtr, nil
}

func (r *roleRepository) Create(ctx context.Context, role *entities.Role) error {
	// Don't cache newly created data to avoid stale cache if transaction rolls back
	return r.repo.Create(ctx, role)
}

func (r *roleRepository) Update(ctx context.Context, role *entities.Role) (int, error) {
	return r.repo.Update(ctx, role)
}

func (r *roleRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	return r.repo.DeleteByID(ctx, id)
}

func (r *roleRepository) Count(ctx context.Context, req *models.GetProjectRolesRequest) (int, error) {
	// Don't cache count queries
	return r.repo.Count(ctx, req)
}

func (r *roleRepository) FindAll(ctx context.Context, req *models.GetProjectRolesRequest) ([]entities.Role, error) {
	// Don't cache paginated lists
	return r.repo.FindAll(ctx, req)
}
