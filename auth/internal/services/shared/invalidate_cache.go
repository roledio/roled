package shared

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
)

func InvalidateAccountCache(ctx context.Context, redis infra.RedisService, accountID string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping account cache invalidation", "account_id", accountID)
		return
	}
	key := rediskeys.AccountByID(accountID)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate account cache", "error", err, "account_id", accountID, "key", key)
	}
}

func InvalidateUserCache(ctx context.Context, redis infra.RedisService, user *entities.User) {
	if user == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping user cache invalidation", "user_id", user.ID)
		return
	}

	// Invalidate user by ID cache
	keys := []string{rediskeys.UserByID(user.ID)}

	// Invalidate user by project and email cache if email is available
	if user.Email != nil {
		keys = append(keys, rediskeys.UserByProjectIDAndEmail(user.ProjectID, *user.Email))
	}

	// Invalidate user by project and external user ID cache if external user ID is available
	if user.ExternalUserID != nil {
		keys = append(keys, rediskeys.UserByProjectIDAndExternalUserID(user.ProjectID, *user.ExternalUserID))
	}

	// Invalidate role by user ID cache
	keys = append(keys, rediskeys.RoleByUserID(user.ID))

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate user cache", "error", err, "user_id", user.ID, "keys", keys)
	}
}

func InvalidateMemberCache(ctx context.Context, redis infra.RedisService, member *entities.Member) {
	if member == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping member cache invalidation", "member_id", member.ID)
		return
	}

	// Invalidate all member cache keys
	keys := []string{
		rediskeys.MemberByID(member.ID),
		rediskeys.MemberByIDJoin(member.ID),
		rediskeys.MemberByAccountIDAndUserID(member.AccountID, member.UserID),
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate member cache", "error", err, "member_id", member.ID, "keys", keys)
	}
}

func InvalidateProjectCache(ctx context.Context, redis infra.RedisService, project *entities.Project) {
	if project == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping project cache invalidation", "project_id", project.ID)
		return
	}

	// Invalidate all project cache keys
	keys := []string{
		rediskeys.ProjectByID(project.ID),
		rediskeys.ProjectByIDAndAccountID(project.ID, project.AccountID),
	}
	if project.IsSystem {
		keys = append(keys, rediskeys.ProjectByIsSystem(true))
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate project cache", "error", err, "project_id", project.ID, "keys", keys)
	}
}

func InvalidateProjectSettingCache(ctx context.Context, redis infra.RedisService, projectID string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping project setting cache invalidation", "project_id", projectID)
		return
	}
	key := rediskeys.ProjectSettingByProjectID(projectID)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate project setting cache", "error", err, "project_id", projectID, "key", key)
	}
}

func InvalidateClientCache(ctx context.Context, redis infra.RedisService, client *entities.Client) {
	if client == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping client cache invalidation", "client_id", client.ID)
		return
	}

	// Invalidate all client cache keys
	keys := []string{
		rediskeys.ClientByID(client.ID),
		rediskeys.ClientByIDAndProjectID(client.ID, client.ProjectID),
	}
	if client.IsDefault {
		keys = append(keys, rediskeys.ClientByProjectIDAndIsDefault(client.ProjectID, true))
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate client cache", "error", err, "client_id", client.ID, "keys", keys)
	}
}

func InvalidateClientPermissionsCache(ctx context.Context, redis infra.RedisService, clientID string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping client permissions cache invalidation", "client_id", clientID)
		return
	}

	// Invalidate client permission cache
	key := rediskeys.PermissionsByClientID(clientID)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate client permissions cache", "error", err, "client_id", clientID, "key", key)
	}
}

func InvalidateAccessTokenCache(ctx context.Context, redis infra.RedisService, tokenID string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping access token cache invalidation", "token_id", tokenID)
		return
	}

	// Invalidate all access token cache keys
	keys := []string{
		rediskeys.AccessTokenByID(tokenID),
		rediskeys.AccessTokenByIDJoin(tokenID),
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate access token cache", "error", err, "token_id", tokenID, "keys", keys)
	}
}

func InvalidateRefreshTokenCache(ctx context.Context, redis infra.RedisService, clientID, tokenHash string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping refresh token cache invalidation", "client_id", clientID)
		return
	}

	// Invalidate refresh token cache
	key := rediskeys.RefreshTokenByClientIDAndTokenHash(clientID, tokenHash)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate refresh token cache", "error", err, "client_id", clientID, "key", key)
	}
}

func InvalidateAuthCodeCache(ctx context.Context, redis infra.RedisService, clientID, codeHash string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping auth code cache invalidation", "client_id", clientID)
		return
	}

	// Invalidate auth code cache
	key := rediskeys.AuthCodeByClientIDAndCodeHash(clientID, codeHash)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate auth code cache", "error", err, "client_id", clientID, "key", key)
	}
}

func InvalidateRoleCache(ctx context.Context, redis infra.RedisService, registry repositories.Registry, role *entities.Role) {
	if role == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping role cache invalidation", "role_id", role.ID)
		return
	}

	// Invalidate all role cache keys
	keys := []string{
		rediskeys.RoleByIDAndProjectID(role.ID, role.ProjectID),
		rediskeys.RoleByProjectIDAndCode(role.ProjectID, role.Code),
	}

	// Invalidate RoleByUserID cache for all users with this role
	userIDs, err := registry.UserRoleRepository().FindUserIDsByRoleID(ctx, role.ID)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to find user IDs by role ID for cache invalidation",
			"error", err,
			"role_id", role.ID)
	} else {
		for _, userID := range userIDs {
			keys = append(keys, rediskeys.RoleByUserID(userID))
		}
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate role cache", "error", err, "role_id", role.ID, "keys", keys)
	}
}

func InvalidateRolePermissionsCache(ctx context.Context, redis infra.RedisService, roleID string) {
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping role permissions cache invalidation", "role_id", roleID)
		return
	}

	// Invalidate role permission cache
	key := rediskeys.PermissionsByRoleID(roleID)
	if err := redis.DeleteManyWithContext(ctx, []string{key}); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate role permissions cache", "error", err, "role_id", roleID, "key", key)
	}
}

func InvalidateResourceCache(ctx context.Context, redis infra.RedisService, resource *entities.Resource) {
	if resource == nil {
		return
	}
	if redis == nil {
		log.WithContext(ctx).Warnw("Redis service is nil, skipping resource cache invalidation", "resource_id", resource.ID)
		return
	}

	// Invalidate all resource cache keys
	keys := []string{
		rediskeys.ResourceByIDAndProjectID(resource.ID, resource.ProjectID),
		rediskeys.ResourceByProjectIDAndCode(resource.ProjectID, resource.Code),
	}

	// Delete all cache keys in a single operation
	if err := redis.DeleteManyWithContext(ctx, keys); err != nil {
		log.WithContext(ctx).Warnw("Failed to invalidate resource cache", "error", err, "resource_id", resource.ID, "keys", keys)
	}
}
