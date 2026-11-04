package oauthconnection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
)

func TestUpdateOAuthConnection_Success_Custom(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	existingConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              stringPtr("old-client"),
		ClientSecretEncrypted: stringPtr("old-secret"),
		Scopes:                stringPtr("openid"),
		Enabled:               true,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(1, nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "new-client",
		ClientSecret:   "new-secret",
		Scopes:         []string{"openid", "profile", "email"},
		Enabled:        boolPtr(false),
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeCustom, result.CredentialType)
	assert.False(t, result.Enabled)
}

func TestUpdateOAuthConnection_Success_Default(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID:       "proj-123",
		IsSystem: false,
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	existingConnection := &entities.OAuthConnection{
		ID:                    "conn-1",
		ProjectID:             "proj-123",
		Provider:              "google",
		CredentialType:        constants.OAuthCredentialTypeCustom,
		ClientID:              stringPtr("old-client"),
		ClientSecretEncrypted: stringPtr("old-secret"),
		Scopes:                stringPtr("openid"),
		Enabled:               true,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)

	// System project lookup required when switching to default credentials
	systemProject := &entities.Project{ID: "proj-system", IsSystem: true}
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)
	systemConn := &entities.OAuthConnection{
		ID:             "sys-conn-1",
		ProjectID:      "proj-system",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		Enabled:        true,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-system", "google").Return(systemConn, nil)

	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(1, nil)

	mockRedisService.EXPECT().
		DeleteManyWithContext(ctx, []string{
			"oauth_connection:project:proj-123",
			"oauth_connection:project:proj-123:provider:google",
		}).
		Return(nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeDefault,
		Enabled:        boolPtr(true),
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "conn-1", result.ID)
	assert.Equal(t, "proj-123", result.ProjectID)
	assert.Equal(t, "google", result.Provider)
	assert.Equal(t, constants.OAuthCredentialTypeDefault, result.CredentialType)
	assert.True(t, result.Enabled)
}

func TestUpdateOAuthConnection_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(nil, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrProjectNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_ConnectionNotFound(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(nil, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_UpdateFails(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	existingConnection := &entities.OAuthConnection{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(0, assert.AnError)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_NoRowsAffected(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

	project := &entities.Project{
		ID: "proj-123",
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

	existingConnection := &entities.OAuthConnection{
		ID:             "conn-1",
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
	}
	mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)
	mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(0, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-123",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeCustom,
		ClientID:       "client-1",
		ClientSecret:   "secret-1",
		Scopes:         []string{"openid"},
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrOAuthConnectionNotFound.Msg, err.Error())
	assert.Nil(t, result)
}

func stringPtr(s string) *string {
	return &s
}

func TestUpdateOAuthConnection_SystemProject_DefaultCredentials_Rejected(t *testing.T) {
	ctx := context.Background()

	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockRedisService := redismocks.NewMockService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	project := &entities.Project{
		ID:       "proj-system",
		IsSystem: true,
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "proj-system").Return(project, nil)

	service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)

	req := &models.UpdateOAuthConnectionRequest{
		ProjectID:      "proj-system",
		Provider:       "google",
		CredentialType: constants.OAuthCredentialTypeDefault,
		Enabled:        boolPtr(true),
	}

	result, err := service.UpdateOAuthConnection(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrDefaultCredentialsNotSupported.Msg, err.Error())
	assert.Nil(t, result)
}

func TestUpdateOAuthConnection_DefaultCredentials_SystemConnectionStates(t *testing.T) {
	testCases := []struct {
		name                    string
		previousCredentialType  string
		systemConnectionExists  bool
		systemConnectionEnabled bool
		expectedError           string
	}{
		{
			name:                   "PreviouslyCustom/SystemConnectionMissing",
			previousCredentialType: constants.OAuthCredentialTypeCustom,
			expectedError:          errors.ErrDefaultCredentialsNotConfigured.Msg,
		},
		{
			name:                   "PreviouslyDefault/SystemConnectionMissing",
			previousCredentialType: constants.OAuthCredentialTypeDefault,
			expectedError:          errors.ErrDefaultCredentialsNotConfigured.Msg,
		},
		{
			name:                    "PreviouslyCustom/SystemConnectionDisabled",
			previousCredentialType:  constants.OAuthCredentialTypeCustom,
			systemConnectionExists:  true,
			systemConnectionEnabled: false,
		},
		{
			name:                    "PreviouslyCustom/SystemConnectionEnabled",
			previousCredentialType:  constants.OAuthCredentialTypeCustom,
			systemConnectionExists:  true,
			systemConnectionEnabled: true,
		},
		{
			name:                    "PreviouslyDefault/SystemConnectionDisabled",
			previousCredentialType:  constants.OAuthCredentialTypeDefault,
			systemConnectionExists:  true,
			systemConnectionEnabled: false,
		},
		{
			name:                    "PreviouslyDefault/SystemConnectionEnabled",
			previousCredentialType:  constants.OAuthCredentialTypeDefault,
			systemConnectionExists:  true,
			systemConnectionEnabled: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			account := &entities.Account{ID: "acc-123", IsSystem: true, IsActive: true}
			ctx = context.WithValue(ctx, constants.CtxAccount, account)

			mockRegistry := repositorymocks.NewMockRegistry(t)
			mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
			mockOAuthConnectionRepo := interfacemocks.NewMockOAuthConnectionRepository(t)
			mockRedisService := redismocks.NewMockService(t)

			mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
			mockRegistry.EXPECT().OAuthConnectionRepository().Return(mockOAuthConnectionRepo)

			project := &entities.Project{ID: "proj-123", IsSystem: false, IsActive: true}
			mockProjectRepo.EXPECT().FindByID(ctx, "proj-123").Return(project, nil)

			existingConnection := &entities.OAuthConnection{
				ID:             "conn-1",
				ProjectID:      "proj-123",
				Provider:       "google",
				CredentialType: testCase.previousCredentialType,
			}
			mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-123", "google").Return(existingConnection, nil)

			systemProject := &entities.Project{ID: "proj-system", IsSystem: true}
			mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

			var systemConnection *entities.OAuthConnection
			if testCase.systemConnectionExists {
				systemConnection = &entities.OAuthConnection{
					ID:             "sys-conn-1",
					ProjectID:      "proj-system",
					Provider:       "google",
					CredentialType: constants.OAuthCredentialTypeCustom,
					Enabled:        testCase.systemConnectionEnabled,
				}
			}
			mockOAuthConnectionRepo.EXPECT().FindByProjectIDAndProvider(ctx, "proj-system", "google").Return(systemConnection, nil)

			if testCase.expectedError == "" {
				mockOAuthConnectionRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entities.OAuthConnection")).Return(1, nil)
				mockRedisService.EXPECT().
					DeleteManyWithContext(ctx, []string{
						"oauth_connection:project:proj-123",
						"oauth_connection:project:proj-123:provider:google",
					}).
					Return(nil)
			}

			service := NewOAuthConnectionService(&configs.DefaultConfig{EncryptionMasterKey: "test-key"}, mockRegistry, mockRedisService)
			req := &models.UpdateOAuthConnectionRequest{
				ProjectID:      "proj-123",
				Provider:       "google",
				CredentialType: constants.OAuthCredentialTypeDefault,
				Enabled:        boolPtr(true),
			}

			result, err := service.UpdateOAuthConnection(ctx, req)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedError, err.Error())
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "conn-1", result.ID)
			assert.Equal(t, constants.OAuthCredentialTypeDefault, result.CredentialType)
			assert.True(t, result.Enabled)
		})
	}
}
