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

type userRepository struct {
	repo  interfaces.UserRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewUserRepository(repo interfaces.UserRepository, redis infra.RedisService, ttl time.Duration) interfaces.UserRepository {
	if redis == nil {
		return repo
	}
	return &userRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	cacheKey := rediskeys.UserByID(id)
	var user entities.User

	found, err := r.redis.GetData(ctx, cacheKey, &user)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get user from redis cache, falling back to DB", "error", err, "user_id", id)
	} else if found {
		return &user, nil
	}

	userPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if userPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, userPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache user in redis", "error", setErr, "user_id", id)
	}

	return userPtr, nil
}

func (r *userRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.User, error) {
	// Try to get from cache first using individual key
	cacheKey := rediskeys.UserByID(id)
	var user entities.User

	found, err := r.redis.GetData(ctx, cacheKey, &user)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get user from redis cache, falling back to DB", "error", err, "user_id", id)
	} else if found {
		return &user, nil
	}

	// Cache miss, fetch from DB
	userPtr, err := r.repo.FindByIDAndProjectID(ctx, id, projectID)
	if err != nil {
		return nil, err
	}
	if userPtr == nil {
		return nil, nil
	}

	// Cache with all available keys
	cacheKeys := []string{
		rediskeys.UserByID(id),
	}
	if userPtr.Email != nil {
		cacheKeys = append(cacheKeys, rediskeys.UserByProjectIDAndEmail(userPtr.ProjectID, *userPtr.Email))
	}
	if userPtr.ExternalUserID != nil {
		cacheKeys = append(cacheKeys, rediskeys.UserByProjectIDAndExternalUserID(userPtr.ProjectID, *userPtr.ExternalUserID))
	}

	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, userPtr, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache user in redis", "error", setErr, "user_id", id)
		}
	}

	return userPtr, nil
}

func (r *userRepository) FindByProjectIDAndEmail(ctx context.Context, projectID, email string) (*entities.User, error) {
	cacheKey := rediskeys.UserByProjectIDAndEmail(projectID, email)
	var user entities.User

	found, err := r.redis.GetData(ctx, cacheKey, &user)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get user from redis cache, falling back to DB", "error", err, "project_id", projectID, "email", email)
	} else if found {
		return &user, nil
	}

	userPtr, err := r.repo.FindByProjectIDAndEmail(ctx, projectID, email)
	if err != nil {
		return nil, err
	}
	if userPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, userPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache user in redis", "error", setErr, "project_id", projectID, "email", email)
	}

	return userPtr, nil
}

func (r *userRepository) FindByProjectIDAndExternalUserID(ctx context.Context, projectID, externalUserID string) (*entities.User, error) {
	cacheKey := rediskeys.UserByProjectIDAndExternalUserID(projectID, externalUserID)
	var user entities.User

	found, err := r.redis.GetData(ctx, cacheKey, &user)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get user from redis cache, falling back to DB", "error", err, "project_id", projectID, "external_user_id", externalUserID)
	} else if found {
		return &user, nil
	}

	userPtr, err := r.repo.FindByProjectIDAndExternalUserID(ctx, projectID, externalUserID)
	if err != nil {
		return nil, err
	}
	if userPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, userPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache user in redis", "error", setErr, "project_id", projectID, "external_user_id", externalUserID)
	}

	return userPtr, nil
}

func (r *userRepository) FindByProjectIDAndExternalUserIDJoinRole(ctx context.Context, projectID, externalUserID string) (*interfaces.UserAndRole, error) {
	// Don't cache join queries
	return r.repo.FindByProjectIDAndExternalUserIDJoinRole(ctx, projectID, externalUserID)
}

func (r *userRepository) FindByIDAndProjectIDJoinRole(ctx context.Context, userID, projectID string) (*interfaces.UserAndRole, error) {
	// Don't cache join queries
	return r.repo.FindByIDAndProjectIDJoinRole(ctx, userID, projectID)
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	// Don't cache newly created data to avoid stale cache if transaction rolls back
	return r.repo.Create(ctx, user)
}

func (r *userRepository) SetEmailVerified(ctx context.Context, userID string) (int, error) {
	return r.repo.SetEmailVerified(ctx, userID)
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) (int, error) {
	return r.repo.UpdatePassword(ctx, userID, passwordHash)
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) (int, error) {
	return r.repo.Update(ctx, user)
}

func (r *userRepository) DeleteByID(ctx context.Context, userID string) (int, error) {
	return r.repo.DeleteByID(ctx, userID)
}

func (r *userRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	// Don't cache delete operations
	return r.repo.DeleteByAccountID(ctx, accountID)
}

func (r *userRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	// Don't cache delete operations
	return r.repo.DeleteByProjectID(ctx, projectID)
}

func (r *userRepository) Count(ctx context.Context, req *models.GetUsersRequest) (int, error) {
	// Don't cache count queries
	return r.repo.Count(ctx, req)
}

func (r *userRepository) FindAll(ctx context.Context, req *models.GetUsersRequest) ([]interfaces.UserAndRole, error) {
	// Don't cache paginated lists
	return r.repo.FindAll(ctx, req)
}
