package client

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
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateClientProjectGuards(t *testing.T) {
	for _, stage := range []string{"missing account", "project error", "foreign project", "system project"} {
		t.Run(stage, func(t *testing.T) {
			reg := rm.NewMockRegistry(t)
			ctx := context.Background()
			expected := domainerrors.ErrCtxAccountNotFound
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				project := &entities.Project{ID: "project", AccountID: "account", IsSystem: stage == "system project"}
				var failure error
				switch stage {
				case "project error":
					failure = errors.New("offline")
					expected = pkgerrors.ErrSystemError
				case "foreign project":
					project = nil
					expected = domainerrors.ErrProjectNotFound
				default:
					expected = pkgerrors.ErrOperationNotAvailable
				}
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(project, failure).Once()
			}
			s := NewClientService(&configs.DefaultConfig{}, reg, nil)
			var err error
			_, err = s.CreateClient(ctx, &models.CreateClientRequest{ProjectID: "project"})
			require.ErrorIs(t, err, expected)
		})
	}
}

func TestCreateClientTransactionAndSecret(t *testing.T) {
	for _, stage := range []string{"success", "no permissions", "permission error", "missing permissions", "create error", "links error", "commit error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("offline")
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			permissions := im.NewMockPermissionRepository(t)
			clients := im.NewMockClientRepository(t)
			links := im.NewMockClientPermissionRepository(t)
			reg.On("ProjectRepository").Return(projects).Once()
			projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
			ids := []string{"permission", "second"}
			if stage == "no permissions" {
				ids = nil
			}
			stop := false
			if len(ids) > 0 {
				reg.On("PermissionRepository").Return(permissions).Once()
				found := []interfaces.PermissionResource{{ID: "permission", Name: "Read", ResourceName: "Records"}, {ID: "second", Name: "Write", ResourceName: "Records"}}
				var e error
				if stage == "permission error" {
					e = failure
				}
				if stage == "missing permissions" {
					found = found[:1]
				}
				permissions.On("FindByIDs", ctx, ids).Return(found, e).Once()
				stop = e != nil || stage == "missing permissions"
			}
			var saved *entities.Client
			master := "12345678901234567890123456789012"
			if !stop {
				tx.On("ClientRepository").Return(clients).Once()
				var createErr error
				if stage == "create error" {
					createErr = failure
				}
				clients.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
					saved = a.Get(1).(*entities.Client)
					require.NotEmpty(t, saved.ID)
					require.Equal(t, "account", saved.AccountID)
					require.Equal(t, "project", saved.ProjectID)
					require.True(t, saved.IsActive)
					require.False(t, saved.IsDefault)
					key, e := encryptionutil.DeriveKey([]byte(master), pkgconstants.KeyPurposeClientSecret)
					require.NoError(t, e)
					secret, e := encryptionutil.DecryptAES(saved.SecretEncrypted, key, pkgconstants.KeyPurposeClientSecret)
					require.NoError(t, e)
					require.Len(t, secret, 64)
				}).Return(createErr).Once()
				if createErr == nil && len(ids) > 0 {
					tx.On("ClientPermissionRepository").Return(links).Once()
					var e error
					if stage == "links error" {
						e = failure
					}
					links.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
						rows := a.Get(1).([]entities.ClientPermission)
						require.Equal(t, []entities.ClientPermission{{ClientID: saved.ID, PermissionID: "permission"}, {ClientID: saved.ID, PermissionID: "second"}}, rows)
					}).Return(e).Once()
				}
				reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
					if e := fn(tx); e != nil {
						return e
					}
					if stage == "commit error" {
						return failure
					}
					return nil
				}).Once()
			}
			s := NewClientService(&configs.DefaultConfig{EncryptionMasterKey: master}, reg, nil)
			result, err := s.CreateClient(ctx, &models.CreateClientRequest{ProjectID: "project", Name: "App", PermissionIDs: ids})
			if stage == "success" || stage == "no permissions" {
				require.NoError(t, err)
				require.Equal(t, saved.ID, result.ID)
				require.Equal(t, "App", result.Name)
				require.Len(t, result.Permissions, len(ids))
				if len(ids) > 0 {
					require.Equal(t, "Read", result.Permissions[0].PermissionName)
					require.Equal(t, "Records", result.Permissions[0].ResourceName)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
