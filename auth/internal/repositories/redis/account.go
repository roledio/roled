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

type accountRepository struct {
	repo  interfaces.AccountRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewAccountRepository(repo interfaces.AccountRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.AccountRepository {
	if redis == nil {
		return repo
	}
	return &accountRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *accountRepository) FindByID(ctx context.Context, id string) (*entities.Account, error) {
	cacheKey := rediskeys.AccountByID(id)
	var account entities.Account

	found, err := r.redis.GetData(ctx, cacheKey, &account)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get account from redis cache, falling back to DB", "error", err, "account_id", id)
	} else if found {
		return &account, nil
	}

	accountPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if accountPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, accountPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache account in redis", "error", setErr, "account_id", id)
	}

	return accountPtr, nil
}

func (r *accountRepository) Create(ctx context.Context, account *entities.Account) error {
	return r.repo.Create(ctx, account)
}

func (r *accountRepository) Update(ctx context.Context, account *entities.Account) (int, error) {
	return r.repo.Update(ctx, account)
}

func (r *accountRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	return r.repo.DeleteByID(ctx, id)
}

func (r *accountRepository) FindAll(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) ([]entities.Account, error) {
	return r.repo.FindAll(ctx, req, filterAccountID)
}

func (r *accountRepository) Count(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) (int, error) {
	return r.repo.Count(ctx, req, filterAccountID)
}
