package project

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	um "github.com/roledio/roled/auth/internal/services/upload/mocks"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectRequiresAccount(t *testing.T) {
	s := &projectService{registry: rm.NewMockRegistry(t)}
	got, err := s.CreateProject(context.Background(), &models.CreateProjectRequest{Name: "Project"})
	require.ErrorIs(t, err, domainerrors.ErrCtxAccountNotFound)
	assert.Nil(t, got)
}

func TestCreateProjectTransaction(t *testing.T) {
	for _, failAt := range []string{"", "project", "redirects", "settings", "client", "resources", "permissions", "client permissions", "upload", "commit"} {
		name := failAt
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			reg := rm.NewMockRegistry(t)
			tx := rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			redirects := im.NewMockRedirectURIRepository(t)
			settings := im.NewMockProjectSettingRepository(t)
			clients := im.NewMockClientRepository(t)
			resources := im.NewMockResourceRepository(t)
			permissions := im.NewMockPermissionRepository(t)
			clientPermissions := im.NewMockClientPermissionRepository(t)
			uploads := um.NewMockUploadService(t)
			failure := errors.New("storage unavailable")
			var savedProject *entities.Project
			var savedClient *entities.Client
			var savedResources []entities.Resource
			var savedPermissions []entities.Permission
			var savedClientPermissions []entities.ClientPermission
			stages := []struct {
				name  string
				setup func(error)
			}{
				{"project", func(err error) {
					tx.EXPECT().ProjectRepository().Return(projects).Once()
					projects.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						savedProject = a.Get(1).(*entities.Project)
						assert.Equal(t, "account", savedProject.AccountID)
						assert.Equal(t, "Project", savedProject.Name)
						assert.True(t, savedProject.IsActive)
						assert.False(t, savedProject.IsSystem)
						require.NotNil(t, savedProject.LogoURL)
						assert.Equal(t, "https://uploads.example.com/logo.png", *savedProject.LogoURL)
					}).Return(err).Once()
				}},
				{"redirects", func(err error) {
					tx.EXPECT().RedirectURIRepository().Return(redirects).Once()
					redirects.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						rows := a.Get(1).([]entities.RedirectURI)
						require.Len(t, rows, 1)
						assert.Equal(t, savedProject.ID, rows[0].ProjectID)
						assert.Equal(t, "https://app.example.com/callback", rows[0].RedirectURI)
						require.NotNil(t, rows[0].LoginURL)
						assert.Equal(t, "https://app.example.com/login", *rows[0].LoginURL)
					}).Return(err).Once()
				}},
				{"settings", func(err error) {
					tx.EXPECT().ProjectSettingRepository().Return(settings).Once()
					settings.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						row := a.Get(1).(*entities.ProjectSetting)
						assert.Equal(t, savedProject.ID, row.ProjectID)
						assert.False(t, row.IsSignupEnabled)
						assert.False(t, row.IsAllowTempEmail)
						assert.Nil(t, row.DefaultSignupRoleID)
					}).Return(err).Once()
				}},
				{"client", func(err error) {
					tx.EXPECT().ClientRepository().Return(clients).Once()
					clients.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						savedClient = a.Get(1).(*entities.Client)
						assert.Equal(t, savedProject.ID, savedClient.ProjectID)
						assert.Equal(t, "account", savedClient.AccountID)
						assert.True(t, savedClient.IsDefault)
						assert.True(t, savedClient.IsActive)
						key, keyErr := encryptionutil.DeriveKey([]byte("test-master-key"), pkgconstants.KeyPurposeClientSecret)
						require.NoError(t, keyErr)
						secret, decryptErr := encryptionutil.DecryptAES(savedClient.SecretEncrypted, key, pkgconstants.KeyPurposeClientSecret)
						require.NoError(t, decryptErr)
						assert.Len(t, secret, 64)
					}).Return(err).Once()
				}},
				{"resources", func(err error) {
					tx.EXPECT().ResourceRepository().Return(resources).Once()
					resources.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) { savedResources = a.Get(1).([]entities.Resource) }).Return(7, err).Once()
				}},
				{"permissions", func(err error) {
					tx.EXPECT().PermissionRepository().Return(permissions).Once()
					permissions.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) { savedPermissions = a.Get(1).([]entities.Permission) }).Return(22, err).Once()
				}},
				{"client permissions", func(err error) {
					tx.EXPECT().ClientPermissionRepository().Return(clientPermissions).Once()
					clientPermissions.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) { savedClientPermissions = a.Get(1).([]entities.ClientPermission) }).Return(err).Once()
				}},
				{"upload", func(err error) { uploads.EXPECT().Move(ctx, "tmp/logo.png", "logo.png").Return(err).Once() }},
			}
			for _, stage := range stages {
				if stage.name == failAt {
					stage.setup(failure)
					break
				}
				stage.setup(nil)
			}
			reg.EXPECT().Tx(mock.Anything).RunAndReturn(func(fn func(repositories.Registry) error) error {
				if err := fn(tx); err != nil {
					return err
				}
				if failAt == "commit" {
					return failure
				}
				return nil
			}).Once()
			s := &projectService{registry: reg, defaultConfig: &configs.DefaultConfig{EncryptionMasterKey: "test-master-key"}, uploadService: uploads, uploadBaseURL: "https://uploads.example.com"}
			logo := "https://uploads.example.com/tmp/logo.png"
			req := &models.CreateProjectRequest{Name: "Project", LogoURL: &logo, RedirectURIs: []models.RedirectURI{
				{RedirectURI: "https://app.example.com/callback", LoginURL: "https://app.example.com/old"},
				{RedirectURI: "https://app.example.com/callback", LoginURL: "https://app.example.com/login"},
			}}
			got, err := s.CreateProject(ctx, req)
			if failAt != "" {
				require.Error(t, err)
				assert.Nil(t, got)
				switch failAt {
				case "upload":
					assert.ErrorIs(t, err, domainerrors.ErrMoveTmpProjectLogo)
				case "commit":
					assert.ErrorIs(t, err, failure)
				default:
					assert.ErrorIs(t, err, pkgerrors.ErrSystemError)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, savedProject.ID, got.ID)
			assert.NotEmpty(t, got.ID)
			assert.Equal(t, req.Name, got.Name)
			assert.True(t, got.IsActive)
			assert.Equal(t, savedProject.LogoURL, got.LogoURL)
			assert.False(t, got.CreatedAt.IsZero())
			assert.Equal(t, []models.RedirectURI{req.RedirectURIs[1]}, got.RedirectURIs)
			require.Len(t, savedResources, 7)
			require.Len(t, savedPermissions, 22)
			require.Len(t, savedClientPermissions, 22)
			resourceIDs := map[string]bool{}
			for _, row := range savedResources {
				assert.Equal(t, "account", row.AccountID)
				assert.Equal(t, savedProject.ID, row.ProjectID)
				assert.True(t, row.IsDefault)
				assert.NotEmpty(t, row.ID)
				assert.False(t, resourceIDs[row.ID])
				resourceIDs[row.ID] = true
			}
			permissionIDs := map[string]bool{}
			for _, row := range savedPermissions {
				assert.True(t, resourceIDs[row.ResourceID])
				assert.True(t, row.IsDefault)
				permissionIDs[row.ID] = true
			}
			for _, row := range savedClientPermissions {
				assert.Equal(t, savedClient.ID, row.ClientID)
				assert.True(t, permissionIDs[row.PermissionID])
			}
		})
	}
}
