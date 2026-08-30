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

type clientRepository struct {
	repo  interfaces.ClientRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewClientRepository(repo interfaces.ClientRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.ClientRepository {
	if redis == nil {
		return repo
	}
	return &clientRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *clientRepository) getCacheKeysForClient(client *entities.Client) []string {
	cacheKeys := []string{
		rediskeys.ClientByID(client.ID),
		rediskeys.ClientByIDAndProjectID(client.ID, client.ProjectID),
	}
	if client.IsDefault {
		cacheKeys = append(cacheKeys, rediskeys.ClientByProjectIDAndIsDefault(client.ProjectID, true))
	}
	return cacheKeys
}

func (r *clientRepository) setClientCaches(ctx context.Context, client *entities.Client) {
	// Set caches with available keys for client
	cacheKeys := r.getCacheKeysForClient(client)
	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, client, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache client in redis",
				"error", setErr,
				"client_id", client.ID,
				"cache_key", key)
		}
	}
}

func (r *clientRepository) FindByID(ctx context.Context, id string) (*entities.Client, error) {
	cacheKey := rediskeys.ClientByID(id)
	var client entities.Client

	found, err := r.redis.GetData(ctx, cacheKey, &client)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get client from redis cache, falling back to DB", "error", err, "client_id", id)
	} else if found {
		return &client, nil
	}

	clientPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if clientPtr == nil {
		return nil, nil
	}

	r.setClientCaches(ctx, clientPtr)

	return clientPtr, nil
}

func (r *clientRepository) FindByProjectIDAndIsDefault(ctx context.Context, projectID string, isDefault bool) (*entities.Client, error) {
	// Try to get from cache first using composite key
	cacheKey := rediskeys.ClientByProjectIDAndIsDefault(projectID, isDefault)
	var client entities.Client

	found, err := r.redis.GetData(ctx, cacheKey, &client)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get client from redis cache, falling back to DB", "error", err, "project_id", projectID, "is_default", isDefault)
	} else if found {
		return &client, nil
	}

	// Cache miss, fetch from DB
	clientPtr, err := r.repo.FindByProjectIDAndIsDefault(ctx, projectID, isDefault)
	if err != nil {
		return nil, err
	}
	if clientPtr == nil {
		return nil, nil
	}

	r.setClientCaches(ctx, clientPtr)

	return clientPtr, nil
}

func (r *clientRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Client, error) {
	// Try to get from cache first using individual key
	cacheKey := rediskeys.ClientByIDAndProjectID(id, projectID)
	var client entities.Client

	found, err := r.redis.GetData(ctx, cacheKey, &client)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get client from redis cache, falling back to DB", "error", err, "client_id", id)
	} else if found {
		return &client, nil
	}

	// Cache miss, fetch from DB
	clientPtr, err := r.repo.FindByIDAndProjectID(ctx, id, projectID)
	if err != nil {
		return nil, err
	}
	if clientPtr == nil {
		return nil, nil
	}

	r.setClientCaches(ctx, clientPtr)

	return clientPtr, nil
}

func (r *clientRepository) Create(ctx context.Context, client *entities.Client) error {
	// Don't cache newly created data to avoid stale cache if transaction rolls back
	return r.repo.Create(ctx, client)
}

func (r *clientRepository) Update(ctx context.Context, client *entities.Client) (int, error) {
	return r.repo.Update(ctx, client)
}

func (r *clientRepository) Delete(ctx context.Context, client *entities.Client) (int, error) {
	return r.repo.Delete(ctx, client)
}

func (r *clientRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	return r.repo.DeleteByProjectID(ctx, projectID)
}

func (r *clientRepository) Count(ctx context.Context, req *models.GetClientsRequest) (int, error) {
	return r.repo.Count(ctx, req)
}

func (r *clientRepository) FindAll(ctx context.Context, req *models.GetClientsRequest) ([]entities.Client, error) {
	return r.repo.FindAll(ctx, req)
}
