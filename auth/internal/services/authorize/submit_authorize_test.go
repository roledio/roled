package authorize

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
)

func TestAuthorizeService_SubmitAuthorize_Success(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Simulate transaction
		return fn(mockRegistry)
	})

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock user found
	passwordPlain := "password"
	passwordHash, err := passwordutil.HashPassword(passwordPlain)
	assert.NoError(t, err)
	user := &entities.User{
		ID:           "test-user",
		AccountID:    "test-account",
		Email:        pointy.String("testuser@yahoo.com"),
		PasswordHash: &passwordHash,
		IsActive:     true,
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "testuser@yahoo.com").Return(user, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock auth code create success
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "testuser@yahoo.com",
		Password: passwordPlain,
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Code)
	assert.Len(t, result.Code, 64) // NanoID 64 chars
}

func TestAuthorizeService_SubmitAuthorize_UserNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock user not found
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "invaliduser@yahoo.com").Return(nil, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "invaliduser@yahoo.com",
		Password: "password",
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidUserCredentials, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_AuthCodeCreateError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Simulate transaction
		return fn(mockRegistry)
	})

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock user found
	passwordPlain := "password"
	passwordHash, err := passwordutil.HashPassword(passwordPlain)
	assert.NoError(t, err)
	user := &entities.User{
		ID:           "test-user",
		AccountID:    "test-account",
		Email:        pointy.String("testuser@yahoo.com"),
		PasswordHash: &passwordHash,
		IsActive:     true,
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "testuser@yahoo.com").Return(user, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock auth code create error
	dbErr := assert.AnError
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(dbErr)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "testuser@yahoo.com",
		Password: passwordPlain,
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "An unexpected error occurred")
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_InvalidPassword(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock user found with wrong password hash
	wrongPasswordHash := "?a?$wronghashforpasswordcheck" // Hash that doesn't match "password"
	user := &entities.User{
		ID:           "test-user",
		AccountID:    "test-account",
		Email:        pointy.String("testuser@yahoo.com"),
		PasswordHash: &wrongPasswordHash,
		IsActive:     true,
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "testuser@yahoo.com").Return(user, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "testuser@yahoo.com",
		Password: "password", // Doesn't match the hash
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidUserCredentials, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_UserHasNoPassword(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock user found but no password
	user := &entities.User{
		ID:           "test-user",
		AccountID:    "test-account",
		Email:        pointy.String("testuser@yahoo.com"),
		PasswordHash: nil, // No password
		IsActive:     true,
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "testuser@yahoo.com").Return(user, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "testuser@yahoo.com",
		Password: "password",
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidUserCredentials, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_UserNotActive(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found
	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock user found but not active
	passwordHash := "?a?$qCeyqwTr4MSKekZLs0iJNOj/mm7kTzD1zk36vSyudSPLh5RAmDcUa"
	user := &entities.User{
		ID:           "test-user",
		AccountID:    "test-account",
		Email:        pointy.String("testuser@yahoo.com"),
		PasswordHash: &passwordHash,
		IsActive:     false, // Not active
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "testuser@yahoo.com").Return(user, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:    "testuser@yahoo.com",
		Password: "password",
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidUserCredentials, err)
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_SignupSuccess_SystemProject(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRoleRepo := repositorymocks.NewMockUserRoleRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockRegistry.EXPECT().UserRoleRepository().Return(mockUserRoleRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).Return(nil).Run(func(fn func(repositories.Registry) error) {
		err := fn(mockRegistry)
		assert.NoError(t, err)
	})

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found (system project)
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
		IsSystem: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock project setting found
	roleID := "default-role"
	projectSetting := &entities.ProjectSetting{
		ProjectID:           "test-project",
		IsSignupEnabled:     true,
		DefaultSignupRoleID: &roleID,
		IsAllowTempEmail:    true,
		IsSignupVerifyEmail: false,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock user not found (for signup)
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "newuser@example.com").Return(nil, nil)

	// Mock account creation
	mockAccountRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.Account")).Return(nil)

	// Mock member creation
	mockMemberRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.Member")).Return(nil)

	// Mock user creation
	mockUserRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.User")).Return(nil)

	// Mock user role creation
	mockUserRoleRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserRole")).Return(nil)

	// Mock auth code creation
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:                "newuser@example.com",
		Password:             "password123",
		PasswordConfirmation: "password123",
		IsSignup:             true,
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Code)
}

func TestAuthorizeService_SubmitAuthorize_SignupNotEnabled(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Mock project setting found (signup not enabled)
	projectSetting := &entities.ProjectSetting{
		ProjectID:       "test-project",
		IsSignupEnabled: false,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:                "newuser@example.com",
		Password:             "password123",
		PasswordConfirmation: "password123",
		IsSignup:             true,
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnableToProcessSignup.Error(), err.Error())
	assert.Nil(t, result)
}

func TestAuthorizeService_SubmitAuthorize_UserAlreadyExists(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock client found
	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	// Mock project found
	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	// Mock project setting found
	roleID := "default-role"
	projectSetting := &entities.ProjectSetting{
		ProjectID:           "test-project",
		IsSignupEnabled:     true,
		DefaultSignupRoleID: &roleID,
		IsAllowTempEmail:    true,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	// Mock user already exists
	existingUser := &entities.User{
		ID:    "existing-user",
		Email: pointy.String("existing@example.com"),
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "existing@example.com").Return(existingUser, nil)

	// Mock account found
	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	// Mock redirect URI found
	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	// Create service
	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	// Test request
	req := &models.SubmitAuthorizeRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
		Email:                "existing@example.com",
		Password:             "password123",
		PasswordConfirmation: "password123",
		IsSignup:             true,
	}

	// Execute
	result, err := service.SubmitAuthorize(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUnableToProcessSignup.Msg, err.Error())
	assert.Nil(t, result)
}
