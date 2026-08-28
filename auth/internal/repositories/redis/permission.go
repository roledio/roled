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

type permissionRepository struct {
	repo  interfaces.PermissionRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewPermissionRepository(repo interfaces.PermissionRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.PermissionRepository {
	return &permissionRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]interfaces.PermissionResource, error) {
	cacheKey := rediskeys.PermissionsByRoleID(roleID)
	var permissions []interfaces.PermissionResource

	found, err := r.redis.GetData(ctx, cacheKey, &permissions)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get permissions by role id from redis cache, falling back to DB",
			"error", err,
			"role_id", roleID,
			"cache_key", cacheKey)
	} else if found {
		return permissions, nil
	}

	permissions, err = r.repo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		if setErr := r.redis.SetData(ctx, cacheKey, permissions, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache permissions by role id to redis",
				"error", setErr,
				"role_id", roleID,
				"cache_key", cacheKey)
		}
	}

	return permissions, nil
}

func (r *permissionRepository) FindByClientID(ctx context.Context, clientID string) ([]interfaces.PermissionResource, error) {
	cacheKey := rediskeys.PermissionsByClientID(clientID)
	var permissions []interfaces.PermissionResource

	found, err := r.redis.GetData(ctx, cacheKey, &permissions)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get permissions by client id from redis cache, falling back to DB",
			"error", err,
			"client_id", clientID,
			"cache_key", cacheKey)
	} else if found {
		return permissions, nil
	}

	permissions, err = r.repo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		if setErr := r.redis.SetData(ctx, cacheKey, permissions, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache permissions by client id to redis",
				"error", setErr,
				"client_id", clientID,
				"cache_key", cacheKey)
		}
	}

	return permissions, nil
}

func (r *permissionRepository) FindByResourceIDsAndSearch(ctx context.Context, resourceIDs []string, search string) ([]entities.Permission, error) {
	return r.repo.FindByResourceIDsAndSearch(ctx, resourceIDs, search)
}

func (r *permissionRepository) Create(ctx context.Context, permissions []entities.Permission) (int, error) {
	return r.repo.Create(ctx, permissions)
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []string) ([]interfaces.PermissionResource, error) {
	return r.repo.FindByIDs(ctx, ids)
}

func (r *permissionRepository) DeleteByResourceID(ctx context.Context, resourceID string) (int, error) {
	return r.repo.DeleteByResourceID(ctx, resourceID)
}

func (r *permissionRepository) FindAll(ctx context.Context, req *models.GetPermissionsRequest) ([]interfaces.PermissionResource, error) {
	return r.repo.FindAll(ctx, req)
}

func (r *permissionRepository) Count(ctx context.Context, req *models.GetPermissionsRequest) (int, error) {
	return r.repo.Count(ctx, req)
}

func (r *permissionRepository) FindByProjectID(ctx context.Context, projectID string, isDefault *bool) ([]entities.Permission, error) {
	cacheKey := rediskeys.PermissionsByProjectIDAndIsDefault(projectID, isDefault)
	var permissions []entities.Permission

	found, err := r.redis.GetData(ctx, cacheKey, &permissions)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get permissions by project id from redis cache, falling back to DB",
			"error", err,
			"project_id", projectID,
			"cache_key", cacheKey)
	} else if found {
		return permissions, nil
	}

	permissions, err = r.repo.FindByProjectID(ctx, projectID, isDefault)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		if setErr := r.redis.SetData(ctx, cacheKey, permissions, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache permissions by project id to redis",
				"error", setErr,
				"project_id", projectID,
				"cache_key", cacheKey)
		}
	}

	return permissions, nil
}
