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

type projectRepository struct {
	repo  interfaces.ProjectRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewProjectCacheRepository(repo interfaces.ProjectRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.ProjectRepository {
	if redis == nil {
		return repo
	}
	return &projectRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *projectRepository) getCacheKeysForProject(project *entities.Project) []string {
	cacheKeys := []string{
		rediskeys.ProjectByID(project.ID),
		rediskeys.ProjectByIDAndAccountID(project.ID, project.AccountID),
	}
	if project.IsSystem {
		cacheKeys = append(cacheKeys, rediskeys.ProjectByIsSystem(true))
	}
	return cacheKeys
}

func (r *projectRepository) setProjectCaches(ctx context.Context, project *entities.Project) {
	// Set caches with available keys for project
	cacheKeys := r.getCacheKeysForProject(project)
	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, project, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache project in redis",
				"error", setErr,
				"project_id", project.ID,
				"cache_key", key)
		}
	}
}

func (r *projectRepository) FindByID(ctx context.Context, id string) (*entities.Project, error) {
	cacheKey := rediskeys.ProjectByID(id)
	var project entities.Project

	found, err := r.redis.GetData(ctx, cacheKey, &project)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get project from redis cache, falling back to DB", "error", err, "project_id", id)
	} else if found {
		return &project, nil
	}

	projectPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if projectPtr == nil {
		return nil, nil
	}

	r.setProjectCaches(ctx, projectPtr)

	return projectPtr, nil
}

func (r *projectRepository) FindByIDAndAccountID(ctx context.Context, id string, accountID string) (*entities.Project, error) {
	cacheKey := rediskeys.ProjectByIDAndAccountID(id, accountID)
	var project entities.Project

	found, err := r.redis.GetData(ctx, cacheKey, &project)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get project from redis cache, falling back to DB", "error", err, "project_id", id)
	} else if found {
		return &project, nil
	}

	projectPtr, err := r.repo.FindByIDAndAccountID(ctx, id, accountID)
	if err != nil {
		return nil, err
	}
	if projectPtr == nil {
		return nil, nil
	}

	r.setProjectCaches(ctx, projectPtr)

	return projectPtr, nil
}

func (r *projectRepository) FindSystem(ctx context.Context) (*entities.Project, error) {
	cacheKey := rediskeys.ProjectByIsSystem(true)
	var project entities.Project

	found, err := r.redis.GetData(ctx, cacheKey, &project)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get system project from redis cache, falling back to DB", "error", err)
	} else if found {
		return &project, nil
	}

	projectPtr, err := r.repo.FindSystem(ctx)
	if err != nil {
		return nil, err
	}
	if projectPtr == nil {
		return nil, nil
	}

	r.setProjectCaches(ctx, projectPtr)

	return projectPtr, nil
}

func (r *projectRepository) FindAll(ctx context.Context, req *models.GetProjectsRequest, accountID string) ([]entities.Project, error) {
	return r.repo.FindAll(ctx, req, accountID)
}

func (r *projectRepository) Count(ctx context.Context, req *models.GetProjectsRequest, accountID string) (int, error) {
	return r.repo.Count(ctx, req, accountID)
}

func (r *projectRepository) Create(ctx context.Context, project *entities.Project) error {
	return r.repo.Create(ctx, project)
}

func (r *projectRepository) Update(ctx context.Context, project *entities.Project) (int, error) {
	return r.repo.Update(ctx, project)
}

func (r *projectRepository) Delete(ctx context.Context, project *entities.Project) (int, error) {
	return r.repo.Delete(ctx, project)
}
