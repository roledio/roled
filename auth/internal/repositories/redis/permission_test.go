package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	"github.com/roledio/roled/auth/internal/repositories/redis"
	redism "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPermissionCacheRoleAndClientLookups(t *testing.T) {
	for _, lookup := range []string{"role", "client"} {
		for _, stage := range []string{"hit", "miss", "read error", "write error", "database error", "empty"} {
			t.Run(lookup+"/"+stage, func(t *testing.T) {
				ctx := context.Background()
				db := im.NewMockPermissionRepository(t)
				cache := redism.NewMockService(t)
				ttl := 7 * time.Minute
				failure := errors.New("offline")
				r := redis.NewPermissionRepository(db, cache, ttl)
				method := "FindByRoleID"
				key := rediskeys.PermissionsByRoleID("id")
				if lookup == "client" {
					method = "FindByClientID"
					key = rediskeys.PermissionsByClientID("id")
				}
				expected := []interfaces.PermissionResource{{ID: "read", ResourceName: "Records", Name: "Read"}}
				if stage == "empty" {
					expected = []interfaces.PermissionResource{}
				}
				var readErr error
				if stage == "read error" {
					readErr = failure
				}
				cache.On("GetData", ctx, key, mock.Anything).Run(func(a mock.Arguments) {
					if stage == "hit" {
						*(a.Get(2).(*[]interfaces.PermissionResource)) = expected
					}
				}).Return(stage == "hit", readErr).Once()
				if stage != "hit" {
					var e error
					if stage == "database error" {
						e = failure
					}
					db.On(method, ctx, "id").Return(expected, e).Once()
					if e == nil && len(expected) > 0 {
						var writeErr error
						if stage == "write error" {
							writeErr = failure
						}
						cache.On("SetData", ctx, key, expected, ttl).Return(writeErr).Once()
					}
				}
				var got []interfaces.PermissionResource
				var err error
				if lookup == "role" {
					got, err = r.FindByRoleID(ctx, "id")
				} else {
					got, err = r.FindByClientID(ctx, "id")
				}
				if stage == "database error" {
					require.ErrorIs(t, err, failure)
					require.Nil(t, got)
				} else {
					require.NoError(t, err)
					require.Equal(t, expected, got)
				}
			})
		}
	}
}

func TestPermissionCacheProjectFilterIsolation(t *testing.T) {
	yes, no := true, false
	filters := []*bool{nil, &yes, &no}
	keys := map[string]bool{}
	for _, filter := range filters {
		for _, project := range []string{"tenant-a", "tenant-b"} {
			key := rediskeys.PermissionsByProjectIDAndIsDefault(project, filter)
			require.False(t, keys[key], "cache keys must distinguish tenant and tri-state filter")
			keys[key] = true
		}
	}
	for i, filter := range filters {
		for _, stage := range []string{"hit", "miss", "read error", "write error", "database error", "empty"} {
			t.Run(stage+"/"+[]string{"all", "default", "custom"}[i], func(t *testing.T) {
				ctx := context.Background()
				db := im.NewMockPermissionRepository(t)
				cache := redism.NewMockService(t)
				ttl := time.Minute
				failure := errors.New("offline")
				r := redis.NewPermissionRepository(db, cache, ttl)
				key := rediskeys.PermissionsByProjectIDAndIsDefault("project", filter)
				expected := []entities.Permission{{ID: "read", ResourceID: "records"}}
				if stage == "empty" {
					expected = nil
				}
				var readErr error
				if stage == "read error" {
					readErr = failure
				}
				cache.On("GetData", ctx, key, mock.Anything).Run(func(a mock.Arguments) {
					if stage == "hit" {
						*(a.Get(2).(*[]entities.Permission)) = expected
					}
				}).Return(stage == "hit", readErr).Once()
				if stage != "hit" {
					var e error
					if stage == "database error" {
						e = failure
					}
					db.On("FindByProjectID", ctx, "project", filter).Return(expected, e).Once()
					if e == nil && len(expected) > 0 {
						var writeErr error
						if stage == "write error" {
							writeErr = failure
						}
						cache.On("SetData", ctx, key, expected, ttl).Return(writeErr).Once()
					}
				}
				got, err := r.FindByProjectID(ctx, "project", filter)
				if stage == "database error" {
					require.ErrorIs(t, err, failure)
					require.Nil(t, got)
				} else {
					require.NoError(t, err)
					require.Equal(t, expected, got)
				}
			})
		}
	}
}

func TestPermissionCacheUncachedOperations(t *testing.T) {
	ctx := context.Background()
	db := im.NewMockPermissionRepository(t)
	cache := redism.NewMockService(t)
	require.Same(t, db, redis.NewPermissionRepository(db, nil, time.Minute))
	r := redis.NewPermissionRepository(db, cache, time.Minute)
	failure := errors.New("storage failure")
	rows := []entities.Permission{{ID: "read", ResourceID: "records"}}
	joined := []interfaces.PermissionResource{{ID: "read", ResourceName: "Records"}}
	req := &models.GetPermissionsRequest{ProjectID: "project"}
	// Writes and arbitrary list queries must not populate caches before a transaction commits.
	db.On("Create", ctx, rows).Return(1, nil).Once()
	n, err := r.Create(ctx, rows)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	db.On("DeleteByResourceID", ctx, "records").Return(0, failure).Once()
	n, err = r.DeleteByResourceID(ctx, "records")
	require.ErrorIs(t, err, failure)
	require.Zero(t, n)
	db.On("FindByIDs", ctx, []string{"read"}).Return(joined, nil).Once()
	got, err := r.FindByIDs(ctx, []string{"read"})
	require.NoError(t, err)
	require.Equal(t, joined, got)
	db.On("FindByResourceIDsAndSearch", ctx, []string{"records"}, "read").Return(rows, nil).Once()
	found, err := r.FindByResourceIDsAndSearch(ctx, []string{"records"}, "read")
	require.NoError(t, err)
	require.Equal(t, rows, found)
	db.On("FindAll", ctx, req).Return(joined, nil).Once()
	got, err = r.FindAll(ctx, req)
	require.NoError(t, err)
	require.Equal(t, joined, got)
	db.On("Count", ctx, req).Return(1, nil).Once()
	n, err = r.Count(ctx, req)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}
