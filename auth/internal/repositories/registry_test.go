package repositories

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRegistry_NoRedis creates a new registry with nil redis
func TestNewRegistry_NoRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}

	mockDB := &sqlx.DB{}

	reg := NewRegistry(config, mockDB)

	assert.NotNil(t, reg)
	registry, ok := reg.(*registry)
	require.True(t, ok)
	assert.Nil(t, registry.redisService)
	assert.Equal(t, config, registry.defaultConfig)
}

// TestNewRegistry_EmptyRedisArg creates a registry with empty redis argument
func TestNewRegistry_EmptyRedisArg(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}

	mockDB := &sqlx.DB{}

	// Pass empty redis variadic
	reg := NewRegistry(config, mockDB)

	assert.NotNil(t, reg)
	registry, ok := reg.(*registry)
	require.True(t, ok)
	assert.Nil(t, registry.redisService)
}

// TestPing_WithRealDB would test successful ping but requires real DB
// This is an integration test scenario, skipped in unit tests

// TestTx_WithRealDB would test transaction flow but requires real DB
// This is an integration test scenario, skipped in unit tests

// TestAccountRepository_WithoutRedis returns mariadb repository
func TestAccountRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.AccountRepository()
	assert.NotNil(t, repo)
}

// TestMemberRepository_WithoutRedis returns mariadb repository
func TestMemberRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.MemberRepository()
	assert.NotNil(t, repo)
}

// TestProjectRepository_WithoutRedis
func TestProjectRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.ProjectRepository()
	assert.NotNil(t, repo)
}

// TestOAuthConnectionRepository_WithoutRedis
func TestOAuthConnectionRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.OAuthConnectionRepository()
	assert.NotNil(t, repo)
}

// TestProjectSettingRepository_WithoutRedis
func TestProjectSettingRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.ProjectSettingRepository()
	assert.NotNil(t, repo)
}

// TestRedirectURIRepository_WithoutRedis
func TestRedirectURIRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.RedirectURIRepository()
	assert.NotNil(t, repo)
}

// TestRoleRepository_WithoutRedis
func TestRoleRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.RoleRepository()
	assert.NotNil(t, repo)
}

// TestAuthCodeRepository_WithoutRedis
func TestAuthCodeRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.AuthCodeRepository()
	assert.NotNil(t, repo)
}

// TestAccessTokenRepository_WithoutRedis
func TestAccessTokenRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.AccessTokenRepository()
	assert.NotNil(t, repo)
}

// TestRefreshTokenRepository_WithoutRedis
func TestRefreshTokenRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.RefreshTokenRepository()
	assert.NotNil(t, repo)
}

// TestUserRepository_WithoutRedis
func TestUserRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.UserRepository()
	assert.NotNil(t, repo)
}

// TestUserRoleRepository_AlwaysMariaDB (no Redis caching)
func TestUserRoleRepository_AlwaysMariaDB(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.UserRoleRepository()
	assert.NotNil(t, repo)
}

// TestResourceRepository_WithoutRedis
func TestResourceRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.ResourceRepository()
	assert.NotNil(t, repo)
}

// TestPermissionRepository_WithoutRedis
func TestPermissionRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.PermissionRepository()
	assert.NotNil(t, repo)
}

// TestRolePermissionRepository_AlwaysMariaDB
func TestRolePermissionRepository_AlwaysMariaDB(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.RolePermissionRepository()
	assert.NotNil(t, repo)
}

// TestClientRepository_WithoutRedis
func TestClientRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.ClientRepository()
	assert.NotNil(t, repo)
}

// TestClientPermissionRepository_AlwaysMariaDB
func TestClientPermissionRepository_AlwaysMariaDB(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.ClientPermissionRepository()
	assert.NotNil(t, repo)
}

// TestUserIdentityRepository_WithoutRedis
func TestUserIdentityRepository_WithoutRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo := reg.UserIdentityRepository()
	assert.NotNil(t, repo)
}

// TestAllRepositoriesAccessible verifies all repositories are accessible
func TestAllRepositoriesAccessible(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	// Verify all repository accessors return non-nil
	assert.NotNil(t, reg.AccountRepository())
	assert.NotNil(t, reg.MemberRepository())
	assert.NotNil(t, reg.ProjectRepository())
	assert.NotNil(t, reg.OAuthConnectionRepository())
	assert.NotNil(t, reg.ProjectSettingRepository())
	assert.NotNil(t, reg.RedirectURIRepository())
	assert.NotNil(t, reg.RoleRepository())
	assert.NotNil(t, reg.AuthCodeRepository())
	assert.NotNil(t, reg.AccessTokenRepository())
	assert.NotNil(t, reg.RefreshTokenRepository())
	assert.NotNil(t, reg.UserRepository())
	assert.NotNil(t, reg.UserRoleRepository())
	assert.NotNil(t, reg.ResourceRepository())
	assert.NotNil(t, reg.PermissionRepository())
	assert.NotNil(t, reg.RolePermissionRepository())
	assert.NotNil(t, reg.ClientRepository())
	assert.NotNil(t, reg.ClientPermissionRepository())
	assert.NotNil(t, reg.UserIdentityRepository())
}

// TestRegistryFieldAssignment tests registry field assignment
func TestRegistryFieldAssignment(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 48 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	// Verify fields are properly assigned
	assert.Equal(t, config, reg.defaultConfig)
	assert.Equal(t, mockDB, reg.qx)
	assert.Nil(t, reg.redisService)
}

// TestRegistry_ImplementsInterface verifies Registry interface is properly implemented
func TestRegistry_ImplementsInterface(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	_ = NewRegistry(config, mockDB)
}

// TestNewRegistry_MultipleConfigs tests creating registries with different configs
func TestNewRegistry_MultipleConfigs(t *testing.T) {
	config1 := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}

	config2 := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 48 * time.Hour,
	}

	mockDB := &sqlx.DB{}

	reg1 := NewRegistry(config1, mockDB)
	reg2 := NewRegistry(config2, mockDB)

	assert.NotNil(t, reg1)
	assert.NotNil(t, reg2)

	registry1 := reg1.(*registry)
	registry2 := reg2.(*registry)

	assert.Equal(t, 24*time.Hour, registry1.defaultConfig.CacheDefaultTTLDuration)
	assert.Equal(t, 48*time.Hour, registry2.defaultConfig.CacheDefaultTTLDuration)
}

// TestAccountRepository_Consistency tests that multiple calls return repositories
func TestAccountRepository_Consistency(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	// Call multiple times - should create new instances each time
	repo1 := reg.AccountRepository()
	repo2 := reg.AccountRepository()

	assert.NotNil(t, repo1)
	assert.NotNil(t, repo2)
	// They may be different instances since factories create new ones each time
}

// TestProjectRepository_Consistency
func TestProjectRepository_Consistency(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	repo1 := reg.ProjectRepository()
	repo2 := reg.ProjectRepository()

	assert.NotNil(t, repo1)
	assert.NotNil(t, repo2)
}

// TestRegistryDefaultConfig tests registry access to default config
func TestRegistryDefaultConfig(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 36 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	// Verify config is accessible via registry internal state
	assert.Equal(t, 36*time.Hour, reg.defaultConfig.CacheDefaultTTLDuration)
}

// TestRepositoryAccessorWithNilRedis verifies behavior with no Redis service
func TestRepositoryAccessorWithNilRedis(t *testing.T) {
	config := &configs.DefaultConfig{
		CacheDefaultTTLDuration: 24 * time.Hour,
	}
	mockDB := &sqlx.DB{}

	reg := &registry{
		defaultConfig: config,
		qx:            mockDB,
		redisService:  nil,
	}

	// When redisService is nil, repositories should return mariadb implementations
	repo := reg.AccessTokenRepository()
	assert.NotNil(t, repo)

	repo2 := reg.RefreshTokenRepository()
	assert.NotNil(t, repo2)

	repo3 := reg.UserIdentityRepository()
	assert.NotNil(t, repo3)
}
