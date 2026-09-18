package accesstoken

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
)

func TestGetCurrentAccessToken_UserTokenWithRole(t *testing.T) {
	ctx := context.Background()

	// Setup access token in context
	userID := "test-user-id"
	accessToken := &entities.AccessToken{
		ID:        "test-token-id",
		ClientID:  "test-client-id",
		UserID:    &userID,
		GrantType: "authorization_code",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockAccessTokenRepo := interfacemocks.NewMockAccessTokenRepository(t)
	mockPermissionRepo := interfacemocks.NewMockPermissionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo)

	// Mock access token join result with role
	roleID := "test-role-id"
	roleCode := "admin"
	roleName := "Administrator"
	roleDescription := "Admin role"
	projectName := "Test Project"
	clientName := "Test Client"
	userDisplayName := "Test User"
	userEmail := "test@example.com"

	issuedAt := time.Now()
	joinResult := &interfaces.AccessTokenJoinResult{
		ID:              "test-token-id",
		IssuedAt:        issuedAt,
		ExpiresIn:       3600,
		ProjectID:       "test-project-id",
		ProjectName:     projectName,
		ClientID:        "test-client-id",
		ClientName:      clientName,
		UserID:          &userID,
		UserDisplayName: &userDisplayName,
		UserEmail:       &userEmail,
		RoleID:          &roleID,
		RoleCode:        &roleCode,
		RoleName:        &roleName,
		RoleDescription: &roleDescription,
	}
	mockAccessTokenRepo.EXPECT().FindByIDJoin(ctx, "test-token-id").Return(joinResult, nil)

	// Mock role permissions
	rolePermissions := []interfaces.PermissionResource{
		{
			ResourceCode: "users",
			Code:         "read",
		},
		{
			ResourceCode: "users",
			Code:         "create",
		},
	}
	mockPermissionRepo.EXPECT().FindByRoleID(ctx, roleID).Return(rolePermissions, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Execute
	result, err := service.GetCurrentAccessToken(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-token-id", result.ID)
	assert.Equal(t, projectName, result.Project.Name)
	assert.Equal(t, clientName, result.Client.Name)
	assert.NotNil(t, result.User)
	assert.Equal(t, userID, result.User.ID)
	assert.NotNil(t, result.Role)
	assert.Equal(t, roleID, result.Role.ID)
	assert.Equal(t, roleCode, result.Role.Code)

	// Verify permissions come from role
	assert.Len(t, result.Permissions, 2)
	assert.Contains(t, result.Permissions, "users:read")
	assert.Contains(t, result.Permissions, "users:create")
}

func TestGetCurrentAccessToken_UserTokenWithoutRole(t *testing.T) {
	ctx := context.Background()

	// Setup access token in context
	userID := "test-user-id"
	accessToken := &entities.AccessToken{
		ID:        "test-token-id",
		ClientID:  "test-client-id",
		UserID:    &userID,
		GrantType: "authorization_code",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockAccessTokenRepo := interfacemocks.NewMockAccessTokenRepository(t)
	mockPermissionRepo := interfacemocks.NewMockPermissionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo)

	// Mock access token join result without role
	projectName := "Test Project"
	clientName := "Test Client"
	userDisplayName := "Test User"
	userEmail := "test@example.com"

	issuedAt := time.Now()
	joinResult := &interfaces.AccessTokenJoinResult{
		ID:              "test-token-id",
		IssuedAt:        issuedAt,
		ExpiresIn:       3600,
		ProjectID:       "test-project-id",
		ProjectName:     projectName,
		ClientID:        "test-client-id",
		ClientName:      clientName,
		UserID:          &userID,
		UserDisplayName: &userDisplayName,
		UserEmail:       &userEmail,
		RoleID:          nil, // No role assigned
		RoleCode:        nil,
		RoleName:        nil,
		RoleDescription: nil,
	}
	mockAccessTokenRepo.EXPECT().FindByIDJoin(ctx, "test-token-id").Return(joinResult, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Execute
	result, err := service.GetCurrentAccessToken(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-token-id", result.ID)
	assert.Equal(t, projectName, result.Project.Name)
	assert.Equal(t, clientName, result.Client.Name)
	assert.NotNil(t, result.User)
	assert.Equal(t, userID, result.User.ID)
	assert.Nil(t, result.Role)

	// Verify no permissions when user has no role
	assert.Len(t, result.Permissions, 0)
}

func TestGetCurrentAccessToken_ClientToken(t *testing.T) {
	ctx := context.Background()

	// Setup access token in context (no UserID = client token)
	accessToken := &entities.AccessToken{
		ID:        "test-token-id",
		ClientID:  "test-client-id",
		UserID:    nil, // Client token
		GrantType: "client_credentials",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockAccessTokenRepo := interfacemocks.NewMockAccessTokenRepository(t)
	mockPermissionRepo := interfacemocks.NewMockPermissionRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
	mockRegistry.EXPECT().PermissionRepository().Return(mockPermissionRepo)

	// Mock access token join result for client
	projectName := "Test Project"
	clientName := "Test Client"

	issuedAt := time.Now()
	joinResult := &interfaces.AccessTokenJoinResult{
		ID:              "test-token-id",
		IssuedAt:        issuedAt,
		ExpiresIn:       3600,
		ProjectID:       "test-project-id",
		ProjectName:     projectName,
		ClientID:        "test-client-id",
		ClientName:      clientName,
		UserID:          nil,
		UserDisplayName: nil,
		UserEmail:       nil,
		RoleID:          nil,
		RoleCode:        nil,
		RoleName:        nil,
		RoleDescription: nil,
	}
	mockAccessTokenRepo.EXPECT().FindByIDJoin(ctx, "test-token-id").Return(joinResult, nil)

	// Mock client permissions
	clientPermissions := []interfaces.PermissionResource{
		{
			ResourceCode: "projects",
			Code:         "read",
		},
		{
			ResourceCode: "clients",
			Code:         "create",
		},
	}
	mockPermissionRepo.EXPECT().FindByClientID(ctx, "test-client-id").Return(clientPermissions, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Execute
	result, err := service.GetCurrentAccessToken(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-token-id", result.ID)
	assert.Equal(t, projectName, result.Project.Name)
	assert.Equal(t, clientName, result.Client.Name)
	assert.Nil(t, result.User)
	assert.Nil(t, result.Role)

	// Verify permissions come from client
	assert.Len(t, result.Permissions, 2)
	assert.Contains(t, result.Permissions, "projects:read")
	assert.Contains(t, result.Permissions, "clients:create")
}

func TestGetCurrentAccessToken_NoAccessTokenInContext(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAccessTokenService(&defaultConfig, mockRegistry, nil)

	// Execute
	result, err := service.GetCurrentAccessToken(ctx)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
}
