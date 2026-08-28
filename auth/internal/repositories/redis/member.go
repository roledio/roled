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

type memberRepository struct {
	repo  interfaces.MemberRepository
	redis infra.RedisService
	ttl   time.Duration
}

func NewMemberRepository(repo interfaces.MemberRepository, redis infra.RedisService,
	ttl time.Duration) interfaces.MemberRepository {
	if redis == nil {
		return repo
	}
	return &memberRepository{
		repo:  repo,
		redis: redis,
		ttl:   ttl,
	}
}

func (r *memberRepository) getCacheKeysForMember(member *entities.Member) []string {
	cacheKeys := []string{
		rediskeys.MemberByAccountIDAndUserID(member.AccountID, member.UserID),
		rediskeys.MemberByID(member.ID),
		rediskeys.MemberByIDJoin(member.ID),
	}
	return cacheKeys
}

func (r *memberRepository) setMemberCaches(ctx context.Context, member *entities.Member) {
	// Set caches with available keys for member
	cacheKeys := r.getCacheKeysForMember(member)
	for _, key := range cacheKeys {
		if setErr := r.redis.SetData(ctx, key, member, r.ttl); setErr != nil {
			log.WithContext(ctx).Warnw("Failed to cache member in redis",
				"error", setErr,
				"member_id", member.ID,
				"cache_key", key)
		}
	}
}

func (r *memberRepository) Create(ctx context.Context, member *entities.Member) error {
	return r.repo.Create(ctx, member)
}

func (r *memberRepository) FindByAccountIDAndUserID(ctx context.Context, accountID string, userID string) (*entities.Member, error) {
	cacheKey := rediskeys.MemberByAccountIDAndUserID(accountID, userID)
	var member entities.Member

	found, err := r.redis.GetData(ctx, cacheKey, &member)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get member from redis cache, falling back to DB", "error", err, "account_id", accountID, "user_id", userID)
	} else if found {
		return &member, nil
	}

	memberPtr, err := r.repo.FindByAccountIDAndUserID(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	if memberPtr == nil {
		return nil, nil
	}

	r.setMemberCaches(ctx, memberPtr)

	return memberPtr, nil
}

func (r *memberRepository) FindAll(ctx context.Context, req *models.GetMembersRequest) ([]interfaces.MemberUser, error) {
	return r.repo.FindAll(ctx, req)
}

func (r *memberRepository) Count(ctx context.Context, req *models.GetMembersRequest) (int, error) {
	return r.repo.Count(ctx, req)
}

func (r *memberRepository) FindByID(ctx context.Context, id string) (*entities.Member, error) {
	cacheKey := rediskeys.MemberByID(id)
	var member entities.Member

	found, err := r.redis.GetData(ctx, cacheKey, &member)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get member from redis cache, falling back to DB", "error", err, "member_id", id)
	} else if found {
		return &member, nil
	}

	memberPtr, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if memberPtr == nil {
		return nil, nil
	}

	r.setMemberCaches(ctx, memberPtr)

	return memberPtr, nil
}

func (r *memberRepository) Delete(ctx context.Context, member *entities.Member) (int, error) {
	return r.repo.Delete(ctx, member)
}

func (r *memberRepository) Update(ctx context.Context, member *entities.Member) (int, error) {
	return r.repo.Update(ctx, member)
}

func (r *memberRepository) CountByAccountID(ctx context.Context, accountID string, isAdmin *bool) (int, error) {
	return r.repo.CountByAccountID(ctx, accountID, isAdmin)
}

func (r *memberRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	return r.repo.DeleteByAccountID(ctx, accountID)
}

func (r *memberRepository) FindByIDJoinUser(ctx context.Context, id string) (*interfaces.MemberUser, error) {
	cacheKey := rediskeys.MemberByIDJoin(id)
	var member interfaces.MemberUser

	found, err := r.redis.GetData(ctx, cacheKey, &member)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to get member from redis cache, falling back to DB", "error", err, "member_id", id)
	} else if found {
		return &member, nil
	}

	memberPtr, err := r.repo.FindByIDJoinUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if memberPtr == nil {
		return nil, nil
	}

	if setErr := r.redis.SetData(ctx, cacheKey, memberPtr, r.ttl); setErr != nil {
		log.WithContext(ctx).Warnw("Failed to cache member (join user) in redis",
			"error", setErr,
			"member_id", memberPtr.ID,
			"cache_key", cacheKey)
	}

	return memberPtr, nil
}
