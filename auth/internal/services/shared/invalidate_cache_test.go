package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/stretchr/testify/mock"
)

func TestInvalidateAccountCache_Success(t *testing.T) {
	ctx := context.Background()
	accountID := "test-account-id"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.AccountByID(accountID)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateAccountCache(ctx, mockRedis, accountID)
}

func TestInvalidateAccountCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	accountID := "test-account-id"

	// Should not panic when redis is nil
	InvalidateAccountCache(ctx, nil, accountID)
}

func TestInvalidateAccountCache_RedisError(t *testing.T) {
	ctx := context.Background()
	accountID := "test-account-id"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.AccountByID(accountID)
	deleteErr := errors.New("redis delete error")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateAccountCache(ctx, mockRedis, accountID)
}

func TestInvalidateUserCache_Success_WithAllOldData(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}
	oldEmail := "old@example.com"
	oldExternalUserID := "ext-user-789"
	oldData := &OldUserCacheKeyParts{
		Email:          &oldEmail,
		ExternalUserID: &oldExternalUserID,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.UserByID(user.ID),
		rediskeys.UserByIDAndProjectID(user.ID, user.ProjectID),
		rediskeys.UserByProjectIDAndEmail(user.ProjectID, oldEmail),
		rediskeys.UserByProjectIDAndExternalUserID(user.ProjectID, oldExternalUserID),
		rediskeys.RoleByUserID(user.ID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateUserCache(ctx, mockRedis, user, oldData)
}

func TestInvalidateUserCache_Success_WithNilUser(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis
	InvalidateUserCache(ctx, nil, nil, nil)
}

func TestInvalidateUserCache_Success_WithoutOldData(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.UserByID(user.ID),
		rediskeys.UserByIDAndProjectID(user.ID, user.ProjectID),
		rediskeys.RoleByUserID(user.ID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateUserCache(ctx, mockRedis, user, nil)
}

func TestInvalidateUserCache_Success_PartialOldData(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}
	oldEmail := "old@example.com"
	oldData := &OldUserCacheKeyParts{
		Email:          &oldEmail,
		ExternalUserID: nil,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.UserByID(user.ID),
		rediskeys.UserByIDAndProjectID(user.ID, user.ProjectID),
		rediskeys.UserByProjectIDAndEmail(user.ProjectID, oldEmail),
		rediskeys.RoleByUserID(user.ID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateUserCache(ctx, mockRedis, user, oldData)
}

func TestInvalidateUserCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}

	// Should not panic when redis is nil
	InvalidateUserCache(ctx, nil, user, nil)
}

func TestInvalidateUserCache_RedisError(t *testing.T) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateUserCache(ctx, mockRedis, user, nil)
}

func TestInvalidateMemberCache_Success(t *testing.T) {
	ctx := context.Background()
	member := &entities.Member{
		ID:        "member-123",
		AccountID: "account-456",
		UserID:    "user-789",
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.MemberByID(member.ID),
		rediskeys.MemberByIDJoin(member.ID),
		rediskeys.MemberByAccountIDAndUserID(member.AccountID, member.UserID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateMemberCache(ctx, mockRedis, member)
}

func TestInvalidateMemberCache_NilMember(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis
	InvalidateMemberCache(ctx, nil, nil)
}

func TestInvalidateMemberCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	member := &entities.Member{
		ID:        "member-123",
		AccountID: "account-456",
		UserID:    "user-789",
	}

	// Should not panic when redis is nil
	InvalidateMemberCache(ctx, nil, member)
}

func TestInvalidateMemberCache_RedisError(t *testing.T) {
	ctx := context.Background()
	member := &entities.Member{
		ID:        "member-123",
		AccountID: "account-456",
		UserID:    "user-789",
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateMemberCache(ctx, mockRedis, member)
}

func TestInvalidateProjectCache_Success_NonSystemProject(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:        "project-123",
		AccountID: "account-456",
		IsSystem:  false,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.ProjectByID(project.ID),
		rediskeys.ProjectByIDAndAccountID(project.ID, project.AccountID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateProjectCache(ctx, mockRedis, project)
}

func TestInvalidateProjectCache_Success_SystemProject(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:        "project-123",
		AccountID: "account-456",
		IsSystem:  true,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.ProjectByID(project.ID),
		rediskeys.ProjectByIDAndAccountID(project.ID, project.AccountID),
		rediskeys.ProjectByIsSystem(true),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateProjectCache(ctx, mockRedis, project)
}

func TestInvalidateProjectCache_NilProject(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis
	InvalidateProjectCache(ctx, nil, nil)
}

func TestInvalidateProjectCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:        "project-123",
		AccountID: "account-456",
		IsSystem:  false,
	}

	// Should not panic when redis is nil
	InvalidateProjectCache(ctx, nil, project)
}

func TestInvalidateProjectCache_RedisError(t *testing.T) {
	ctx := context.Background()
	project := &entities.Project{
		ID:        "project-123",
		AccountID: "account-456",
		IsSystem:  false,
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateProjectCache(ctx, mockRedis, project)
}

func TestInvalidateRedirectURICache_Success_MultipleURIs(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"
	oldRedirectURIs := []entities.RedirectURI{
		{RedirectURI: "https://old1.com/callback"},
		{RedirectURI: "https://old2.com/callback"},
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.RedirectURIsByProjectID(projectID),
		rediskeys.RedirectURIByProjectIDAndURI(projectID, "https://old1.com/callback"),
		rediskeys.RedirectURIByProjectIDAndURI(projectID, "https://old2.com/callback"),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateRedirectURICache(ctx, mockRedis, projectID, oldRedirectURIs)
}

func TestInvalidateRedirectURICache_Success_NoURIs(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.RedirectURIsByProjectID(projectID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateRedirectURICache(ctx, mockRedis, projectID, []entities.RedirectURI{})
}

func TestInvalidateRedirectURICache_RedisNil(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"

	// Should not panic when redis is nil
	InvalidateRedirectURICache(ctx, nil, projectID, []entities.RedirectURI{})
}

func TestInvalidateRedirectURICache_RedisError(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"
	oldRedirectURIs := []entities.RedirectURI{
		{RedirectURI: "https://old.com/callback"},
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateRedirectURICache(ctx, mockRedis, projectID, oldRedirectURIs)
}

func TestInvalidateProjectSettingCache_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.ProjectSettingByProjectID(projectID)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateProjectSettingCache(ctx, mockRedis, projectID)
}

func TestInvalidateProjectSettingCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"

	// Should not panic when redis is nil
	InvalidateProjectSettingCache(ctx, nil, projectID)
}

func TestInvalidateProjectSettingCache_RedisError(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateProjectSettingCache(ctx, mockRedis, projectID)
}

func TestInvalidateClientCache_Success_NonDefaultClient(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		ProjectID: "project-456",
		IsDefault: false,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.ClientByID(client.ID),
		rediskeys.ClientByIDAndProjectID(client.ID, client.ProjectID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateClientCache(ctx, mockRedis, client)
}

func TestInvalidateClientCache_Success_DefaultClient(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		ProjectID: "project-456",
		IsDefault: true,
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.ClientByID(client.ID),
		rediskeys.ClientByIDAndProjectID(client.ID, client.ProjectID),
		rediskeys.ClientByProjectIDAndIsDefault(client.ProjectID, true),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateClientCache(ctx, mockRedis, client)
}

func TestInvalidateClientCache_NilClient(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis
	InvalidateClientCache(ctx, nil, nil)
}

func TestInvalidateClientCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		ProjectID: "project-456",
		IsDefault: false,
	}

	// Should not panic when redis is nil
	InvalidateClientCache(ctx, nil, client)
}

func TestInvalidateClientCache_RedisError(t *testing.T) {
	ctx := context.Background()
	client := &entities.Client{
		ID:        "client-123",
		ProjectID: "project-456",
		IsDefault: false,
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateClientCache(ctx, mockRedis, client)
}

func TestInvalidateClientPermissionsCache_Success(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.PermissionsByClientID(clientID)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateClientPermissionsCache(ctx, mockRedis, clientID)
}

func TestInvalidateClientPermissionsCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"

	// Should not panic when redis is nil
	InvalidateClientPermissionsCache(ctx, nil, clientID)
}

func TestInvalidateClientPermissionsCache_RedisError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateClientPermissionsCache(ctx, mockRedis, clientID)
}

func TestInvalidateAccessTokenCache_Success(t *testing.T) {
	ctx := context.Background()
	tokenID := "token-123"

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.AccessTokenByID(tokenID),
		rediskeys.AccessTokenByIDJoin(tokenID),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateAccessTokenCache(ctx, mockRedis, tokenID)
}

func TestInvalidateAccessTokenCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	tokenID := "token-123"

	// Should not panic when redis is nil
	InvalidateAccessTokenCache(ctx, nil, tokenID)
}

func TestInvalidateAccessTokenCache_RedisError(t *testing.T) {
	ctx := context.Background()
	tokenID := "token-123"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateAccessTokenCache(ctx, mockRedis, tokenID)
}

func TestInvalidateRefreshTokenCache_Success(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	tokenHash := "hash-abc123"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.RefreshTokenByClientIDAndTokenHash(clientID, tokenHash)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateRefreshTokenCache(ctx, mockRedis, clientID, tokenHash)
}

func TestInvalidateRefreshTokenCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	tokenHash := "hash-abc123"

	// Should not panic when redis is nil
	InvalidateRefreshTokenCache(ctx, nil, clientID, tokenHash)
}

func TestInvalidateRefreshTokenCache_RedisError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	tokenHash := "hash-abc123"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateRefreshTokenCache(ctx, mockRedis, clientID, tokenHash)
}

func TestInvalidateAuthCodeCache_Success(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	codeHash := "hash-xyz789"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.AuthCodeByClientIDAndCodeHash(clientID, codeHash)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateAuthCodeCache(ctx, mockRedis, clientID, codeHash)
}

func TestInvalidateAuthCodeCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	codeHash := "hash-xyz789"

	// Should not panic when redis is nil
	InvalidateAuthCodeCache(ctx, nil, clientID, codeHash)
}

func TestInvalidateAuthCodeCache_RedisError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-123"
	codeHash := "hash-xyz789"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateAuthCodeCache(ctx, mockRedis, clientID, codeHash)
}

func TestInvalidateRoleCache_Success_WithUsers(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}
	userIDs := []string{"user-1", "user-2"}

	mockRedis := redismocks.NewMockService(t)
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(t)

	mockRegistry.EXPECT().
		UserRoleRepository().
		Return(mockUserRoleRepo)

	mockUserRoleRepo.EXPECT().
		FindUserIDsByRoleID(ctx, role.ID).
		Return(userIDs, nil)

	expectedKeys := []string{
		rediskeys.RoleByIDAndProjectID(role.ID, role.ProjectID),
		rediskeys.RoleByProjectIDAndCode(role.ProjectID, role.Code),
		rediskeys.RoleByUserID("user-1"),
		rediskeys.RoleByUserID("user-2"),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateRoleCache(ctx, mockRedis, mockRegistry, role)
}

func TestInvalidateRoleCache_Success_NoUsers(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}

	mockRedis := redismocks.NewMockService(t)
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(t)

	mockRegistry.EXPECT().
		UserRoleRepository().
		Return(mockUserRoleRepo)

	mockUserRoleRepo.EXPECT().
		FindUserIDsByRoleID(ctx, role.ID).
		Return([]string{}, nil)

	expectedKeys := []string{
		rediskeys.RoleByIDAndProjectID(role.ID, role.ProjectID),
		rediskeys.RoleByProjectIDAndCode(role.ProjectID, role.Code),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateRoleCache(ctx, mockRedis, mockRegistry, role)
}

func TestInvalidateRoleCache_NilRole(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis or registry
	InvalidateRoleCache(ctx, nil, nil, nil)
}

func TestInvalidateRoleCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}

	// Should not panic when redis is nil
	InvalidateRoleCache(ctx, nil, nil, role)
}

func TestInvalidateRoleCache_FindUserIDsError(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}

	mockRedis := redismocks.NewMockService(t)
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(t)

	mockRegistry.EXPECT().
		UserRoleRepository().
		Return(mockUserRoleRepo)

	mockUserRoleRepo.EXPECT().
		FindUserIDsByRoleID(ctx, role.ID).
		Return(nil, errors.New("db error"))

	expectedKeys := []string{
		rediskeys.RoleByIDAndProjectID(role.ID, role.ProjectID),
		rediskeys.RoleByProjectIDAndCode(role.ProjectID, role.Code),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	// Should log error but continue with cache invalidation for the role itself
	InvalidateRoleCache(ctx, mockRedis, mockRegistry, role)
}

func TestInvalidateRoleCache_DeleteError(t *testing.T) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}
	userIDs := []string{"user-1"}

	mockRedis := redismocks.NewMockService(t)
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(t)

	mockRegistry.EXPECT().
		UserRoleRepository().
		Return(mockUserRoleRepo)

	mockUserRoleRepo.EXPECT().
		FindUserIDsByRoleID(ctx, role.ID).
		Return(userIDs, nil)

	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateRoleCache(ctx, mockRedis, mockRegistry, role)
}

func TestInvalidateRolePermissionsCache_Success(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"

	mockRedis := redismocks.NewMockService(t)
	expectedKey := rediskeys.PermissionsByRoleID(roleID)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, []string{expectedKey}).
		Return(nil)

	InvalidateRolePermissionsCache(ctx, mockRedis, roleID)
}

func TestInvalidateRolePermissionsCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"

	// Should not panic when redis is nil
	InvalidateRolePermissionsCache(ctx, nil, roleID)
}

func TestInvalidateRolePermissionsCache_RedisError(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateRolePermissionsCache(ctx, mockRedis, roleID)
}

func TestInvalidateResourceCache_Success(t *testing.T) {
	ctx := context.Background()
	resource := &entities.Resource{
		ID:        "resource-123",
		ProjectID: "project-456",
		Code:      "users",
	}

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.ResourceByIDAndProjectID(resource.ID, resource.ProjectID),
		rediskeys.ResourceByProjectIDAndCode(resource.ProjectID, resource.Code),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateResourceCache(ctx, mockRedis, resource)
}

func TestInvalidateResourceCache_NilResource(t *testing.T) {
	ctx := context.Background()

	// Should return early without calling redis
	InvalidateResourceCache(ctx, nil, nil)
}

func TestInvalidateResourceCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	resource := &entities.Resource{
		ID:        "resource-123",
		ProjectID: "project-456",
		Code:      "users",
	}

	// Should not panic when redis is nil
	InvalidateResourceCache(ctx, nil, resource)
}

func TestInvalidateResourceCache_RedisError(t *testing.T) {
	ctx := context.Background()
	resource := &entities.Resource{
		ID:        "resource-123",
		ProjectID: "project-456",
		Code:      "users",
	}

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateResourceCache(ctx, mockRedis, resource)
}

func TestInvalidateOAuthConnectionCache_Success(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"
	provider := "google"

	mockRedis := redismocks.NewMockService(t)

	expectedKeys := []string{
		rediskeys.OAuthConnectionsByProjectID(projectID),
		rediskeys.OAuthConnectionByProjectIDAndProvider(projectID, provider),
	}

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, expectedKeys).
		Return(nil)

	InvalidateOAuthConnectionCache(ctx, mockRedis, projectID, provider)
}

func TestInvalidateOAuthConnectionCache_RedisNil(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"
	provider := "google"

	// Should not panic when redis is nil
	InvalidateOAuthConnectionCache(ctx, nil, projectID, provider)
}

func TestInvalidateOAuthConnectionCache_RedisError(t *testing.T) {
	ctx := context.Background()
	projectID := "project-123"
	provider := "google"

	mockRedis := redismocks.NewMockService(t)
	deleteErr := errors.New("redis delete failed")
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(deleteErr)

	// Should log error but not panic
	InvalidateOAuthConnectionCache(ctx, mockRedis, projectID, provider)
}

// Benchmark tests
func BenchmarkInvalidateAccountCache(b *testing.B) {
	ctx := context.Background()
	accountID := "test-account-id"
	mockRedis := redismocks.NewMockService(b)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(nil).
		Maybe()

	for b.Loop() {
		InvalidateAccountCache(ctx, mockRedis, accountID)
	}
}

func BenchmarkInvalidateUserCache_WithData(b *testing.B) {
	ctx := context.Background()
	user := &entities.User{
		ID:        "user-123",
		ProjectID: "project-456",
	}
	oldEmail := "old@example.com"
	oldData := &OldUserCacheKeyParts{
		Email: &oldEmail,
	}
	mockRedis := redismocks.NewMockService(b)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(nil).
		Maybe()

	for b.Loop() {
		InvalidateUserCache(ctx, mockRedis, user, oldData)
	}
}

func BenchmarkInvalidateProjectCache(b *testing.B) {
	ctx := context.Background()
	project := &entities.Project{
		ID:        "project-123",
		AccountID: "account-456",
		IsSystem:  true,
	}
	mockRedis := redismocks.NewMockService(b)
	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(nil).
		Maybe()

	for b.Loop() {
		InvalidateProjectCache(ctx, mockRedis, project)
	}
}

func BenchmarkInvalidateRoleCache(b *testing.B) {
	ctx := context.Background()
	role := &entities.Role{
		ID:        "role-123",
		ProjectID: "project-456",
		Code:      "admin",
	}
	mockRedis := redismocks.NewMockService(b)
	mockRegistry := repositorymocks.NewMockRegistry(b)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(b)

	mockRegistry.EXPECT().
		UserRoleRepository().
		Return(mockUserRoleRepo).
		Maybe()

	mockUserRoleRepo.EXPECT().
		FindUserIDsByRoleID(ctx, role.ID).
		Return([]string{"user-1", "user-2"}, nil).
		Maybe()

	mockRedis.EXPECT().
		DeleteManyWithContext(ctx, mock.AnythingOfType("[]string")).
		Return(nil).
		Maybe()

	for b.Loop() {
		InvalidateRoleCache(ctx, mockRedis, mockRegistry, role)
	}
}
