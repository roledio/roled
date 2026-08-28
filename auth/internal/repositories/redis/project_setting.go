package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/roledio/roled/internal/services/infra"
)

type projectSettingRepository struct {
	repo  interfaces.ProjectSettingRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewProjectSettingRepository(repo interfaces.ProjectSettingRepository,
	redis infra.RedisService, ttl time.Duration) interfaces.ProjectSettingRepository {
	if redis == nil {
		return repo
	}
	return &projectSettingRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *projectSettingRepository) FindByProjectID(ctx context.Context, projectID string) (*entities.ProjectSetting, error) {
	cacheKey := rediskeys.ProjectSettingByProjectID(projectID)
	var setting entities.ProjectSetting

	found, err := r.redis.GetData(ctx, cacheKey, &setting)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get project setting from redis cache, falling back to DB", "error", err, "project_id", projectID)
	} else if found {
		return &setting, nil
	}

	settingPtr, err := r.repo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if settingPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, settingPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache project setting in redis", "error", setErr, "project_id", projectID)
	}

	return settingPtr, nil
}

func (r *projectSettingRepository) Create(ctx context.Context, projectSetting *entities.ProjectSetting) error {
	return r.repo.Create(ctx, projectSetting)
}

func (r *projectSettingRepository) Update(ctx context.Context, projectSetting *entities.ProjectSetting) (int, error) {
	return r.repo.Update(ctx, projectSetting)
}
