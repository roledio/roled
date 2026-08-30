package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/repositories/redis"
)

func TestCachedMemberRepository_Create(t *testing.T) {
	ctx := context.Background()
	member := &entities.Member{
		ID:        "member-123",
		AccountID: "acc-123",
		UserID:    "user-123",
		IsAdmin:   false,
	}

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Create(ctx, member).
		Return(nil)

	err := cachedRepo.Create(ctx, member)

	assert.NoError(t, err)
}

func TestCachedMemberRepository_FindByAccountIDAndUserID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	accountID := "acc-123"
	userID := "user-123"
	cacheKey := rediskeys.MemberByAccountIDAndUserID(accountID, userID)
	expectedMember := &entities.Member{
		ID:        "member-123",
		AccountID: accountID,
		UserID:    userID,
		IsAdmin:   false,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Member")).
		Run(func(ctx context.Context, key string, dest any) {
			m := dest.(*entities.Member)
			*m = *expectedMember
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByAccountIDAndUserID(ctx, accountID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMember.ID, result.ID)
}

func TestCachedMemberRepository_FindByAccountIDAndUserID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbMember := &entities.Member{
		ID:        "member-123",
		AccountID: "acc-123",
		UserID:    "user-123",
		IsAdmin:   false,
	}

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	accountID := "acc-123"
	userID := "user-123"
	cacheKey := rediskeys.MemberByAccountIDAndUserID(accountID, userID)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Member")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByAccountIDAndUserID(ctx, accountID, userID).
		Return(dbMember, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbMember, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, rediskeys.MemberByID(dbMember.ID), dbMember, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, rediskeys.MemberByIDJoin(dbMember.ID), dbMember, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByAccountIDAndUserID(ctx, accountID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbMember.ID, result.ID)
}

func TestCachedMemberRepository_FindAll_NoCache(t *testing.T) {
	ctx := context.Background()
	expectedMembers := []interfaces.MemberUser{
		{Member: entities.Member{ID: "m1"}, Email: "user1@test.com"},
		{Member: entities.Member{ID: "m2"}, Email: "user2@test.com"},
	}

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		FindAll(ctx, mock.AnythingOfType("*models.GetMembersRequest")).
		Return(expectedMembers, nil)

	req := &models.GetMembersRequest{}
	result, err := cachedRepo.FindAll(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedMembers), len(result))
}

func TestCachedMemberRepository_Count_NoCache(t *testing.T) {
	ctx := context.Background()

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		Count(ctx, mock.AnythingOfType("*models.GetMembersRequest")).
		Return(5, nil)

	req := &models.GetMembersRequest{}
	result, err := cachedRepo.Count(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 5, result)
}

func TestCachedMemberRepository_FindByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	memberID := "member-123"
	cacheKey := rediskeys.MemberByID(memberID)
	expectedMember := &entities.Member{
		ID:        memberID,
		AccountID: "acc-123",
		UserID:    "user-123",
		IsAdmin:   false,
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Member")).
		Run(func(ctx context.Context, key string, dest any) {
			m := dest.(*entities.Member)
			*m = *expectedMember
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByID(ctx, memberID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMember.ID, result.ID)
}

func TestCachedMemberRepository_FindByID_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbMember := &entities.Member{
		ID:        "member-123",
		AccountID: "acc-123",
		UserID:    "user-123",
		IsAdmin:   false,
	}

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	memberID := "member-123"
	cacheKey := rediskeys.MemberByID(memberID)

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*entities.Member")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByID(ctx, memberID).
		Return(dbMember, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, dbMember, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, rediskeys.MemberByAccountIDAndUserID(dbMember.AccountID, dbMember.UserID), dbMember, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, rediskeys.MemberByIDJoin(dbMember.ID), dbMember, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByID(ctx, memberID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbMember.ID, result.ID)
}

func TestCachedMemberRepository_Delete(t *testing.T) {
	ctx := context.Background()
	memberID := "member-123"
	accountID := "acc-123"
	userID := "user-123"

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	targetMember := &entities.Member{
		ID:        memberID,
		AccountID: accountID,
		UserID:    userID}

	mockDBRepo.EXPECT().
		Delete(ctx, targetMember).
		Return(1, nil)

	affected, err := cachedRepo.Delete(ctx, targetMember)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedMemberRepository_Update(t *testing.T) {
	ctx := context.Background()
	memberID := "member-123"
	accountID := "acc-123"
	userID := "user-123"
	isAdmin := true

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	targetMember := &entities.Member{
		ID:        memberID,
		AccountID: accountID,
		UserID:    userID,
		IsAdmin:   isAdmin}

	mockDBRepo.EXPECT().
		Update(ctx, targetMember).
		Return(1, nil)

	affected, err := cachedRepo.Update(ctx, targetMember)

	assert.NoError(t, err)
	assert.Equal(t, 1, affected)
}

func TestCachedMemberRepository_CountByAccountID_NoCache(t *testing.T) {
	ctx := context.Background()
	accountID := "acc-123"

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		CountByAccountID(ctx, accountID, (*bool)(nil)).
		Return(3, nil)

	result, err := cachedRepo.CountByAccountID(ctx, accountID, nil)

	assert.NoError(t, err)
	assert.Equal(t, 3, result)
}

func TestCachedMemberRepository_DeleteByAccountID_NoCache(t *testing.T) {
	ctx := context.Background()
	accountID := "acc-123"

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	mockDBRepo.EXPECT().
		DeleteByAccountID(ctx, accountID).
		Return(3, nil)

	affected, err := cachedRepo.DeleteByAccountID(ctx, accountID)

	assert.NoError(t, err)
	assert.Equal(t, 3, affected)
}

func TestCachedMemberRepository_FindByIDJoinUser_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	memberID := "member-123"
	cacheKey := rediskeys.MemberByIDJoin(memberID)
	expectedMember := &interfaces.MemberUser{
		Member: entities.Member{
			ID:        memberID,
			AccountID: "acc-123",
			UserID:    "user-123",
			IsAdmin:   false,
		},
	}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*interfaces.MemberUser")).
		Run(func(ctx context.Context, key string, dest any) {
			m := dest.(*interfaces.MemberUser)
			*m = *expectedMember
		}).
		Return(true, nil)

	result, err := cachedRepo.FindByIDJoinUser(ctx, memberID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMember.ID, result.ID)
}

func TestCachedMemberRepository_FindByIDJoinUser_CacheMiss(t *testing.T) {
	ctx := context.Background()
	dbMember := &entities.Member{
		ID:        "member-123",
		AccountID: "acc-123",
		UserID:    "user-123",
		IsAdmin:   false,
	}

	mockDBRepo := repositorymocks.NewMockMemberRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	cachedRepo := redis.NewMemberRepository(mockDBRepo, mockRedis, 24*time.Hour)

	memberID := "member-123"
	cacheKey := rediskeys.MemberByIDJoin(memberID)
	expectedMember := &interfaces.MemberUser{Member: *dbMember}

	mockRedis.EXPECT().
		GetData(ctx, cacheKey, mock.AnythingOfType("*interfaces.MemberUser")).
		Return(false, nil)

	mockDBRepo.EXPECT().
		FindByIDJoinUser(ctx, memberID).
		Return(expectedMember, nil)

	mockRedis.EXPECT().
		SetData(ctx, cacheKey, expectedMember, 24*time.Hour).
		Return(nil)

	mockRedis.EXPECT().
		SetData(ctx, rediskeys.MemberByIDJoin(dbMember.ID), expectedMember, 24*time.Hour).
		Return(nil)

	result, err := cachedRepo.FindByIDJoinUser(ctx, memberID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dbMember.ID, result.ID)
}
