package client

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	redism "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClientMutationProjectGuards(t *testing.T) {
	for _, op := range []string{"create", "update", "delete"} {
		for _, stage := range []string{"missing account", "project error", "foreign project", "system project"} {
			t.Run(op+"/"+stage, func(t *testing.T) {
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
				switch op {
				case "create":
					_, err = s.CreateClient(ctx, &models.CreateClientRequest{ProjectID: "project"})
				case "update":
					_, err = s.UpdateClient(ctx, &models.UpdateClientRequest{ProjectID: "project", ClientID: "client"})
				default:
					err = s.DeleteClient(ctx, &models.DeleteClientRequest{ProjectID: "project", ClientID: "client"})
				}
				require.ErrorIs(t, err, expected)
			})
		}
	}
}

func TestUpdateClientPermissionsAndFailures(t *testing.T) {
	for _, stage := range []string{"custom", "default", "empty", "nil active", "lookup error", "missing client", "permission error", "missing permissions", "update error", "zero update", "default permission error", "delete permission error", "create permission error", "commit error", "response error", "cache error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("storage unavailable")
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			clients := im.NewMockClientRepository(t)
			txClients := im.NewMockClientRepository(t)
			permissions := im.NewMockPermissionRepository(t)
			links := im.NewMockClientPermissionRepository(t)
			cache := redism.NewMockService(t)
			reg.On("ProjectRepository").Return(projects).Once()
			projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project", AccountID: "account"}, nil).Once()
			reg.On("ClientRepository").Return(clients).Once()
			stored := &entities.Client{ID: "client", ProjectID: "project", Name: "Old", IsDefault: stage == "default" || stage == "default permission error"}
			ids := []string{"default", "custom"}
			if stage == "empty" {
				ids = nil
			}
			active := true
			req := &models.UpdateClientRequest{ProjectID: "project", ClientID: "client", Name: "Updated", PermissionIDs: ids, IsActive: &active}
			if stage == "nil active" {
				req.IsActive = nil
			}
			var lookupErr error
			if stage == "lookup error" {
				lookupErr = failure
			}
			if stage == "missing client" {
				stored = nil
			}
			clients.On("FindByIDAndProjectID", ctx, "client", "project").Return(stored, lookupErr).Once()
			stop := stage == "lookup error" || stage == "missing client"
			if !stop {
				reg.On("PermissionRepository").Return(permissions).Maybe()
				if len(ids) > 0 {
					found := []interfaces.PermissionResource{{ID: "default", IsDefault: true}, {ID: "custom", Name: "Read", ResourceName: "Records"}}
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
			}
			committed := false
			if !stop {
				tx.On("ClientRepository").Return(txClients).Once()
				n := 1
				var updateErr error
				if stage == "update error" {
					updateErr = failure
				}
				if stage == "zero update" {
					n = 0
				}
				txClients.On("Update", ctx, stored).Run(func(a mock.Arguments) {
					c := a.Get(1).(*entities.Client)
					require.Equal(t, req.Name, c.Name)
					require.Equal(t, req.IsActive != nil && *req.IsActive, c.IsActive)
					require.Nil(t, c.Description)
				}).Return(n, updateErr).Once()
				stop = updateErr != nil || n == 0
				if !stop {
					tx.On("ClientPermissionRepository").Return(links).Once()
					if stored.IsDefault {
						tx.On("PermissionRepository").Return(permissions).Once()
						var e error
						if stage == "default permission error" {
							e = failure
						}
						permissions.On("FindByProjectID", ctx, "project", mock.MatchedBy(func(v *bool) bool { return v != nil && *v })).Return([]entities.Permission{{ID: "default"}}, e).Once()
						stop = e != nil
					}
				}
				if !stop {
					var e error
					if stage == "delete permission error" {
						e = failure
					}
					links.On("DeleteByClientID", ctx, "client").Return(2, e).Once()
					stop = e != nil
				}
				if !stop {
					expected := []entities.ClientPermission{}
					for _, id := range ids {
						expected = append(expected, entities.ClientPermission{ClientID: "client", PermissionID: id})
					}
					var e error
					if stage == "create permission error" {
						e = failure
					}
					links.On("Create", ctx, expected).Return(e).Once()
					stop = e != nil
				}
				reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
					e := fn(tx)
					if e != nil {
						return e
					}
					if stage == "commit error" {
						return failure
					}
					committed = true
					return nil
				}).Once()
				if !stop && stage != "commit error" {
					keys := []string{rediskeys.ClientByID("client"), rediskeys.ClientByIDAndProjectID("client", "project")}
					if stored.IsDefault {
						keys = append(keys, rediskeys.ClientByProjectIDAndIsDefault("project", true))
					}
					var cacheErr error
					if stage == "cache error" {
						cacheErr = failure
					}
					cache.On("DeleteManyWithContext", ctx, keys).Run(func(mock.Arguments) { require.True(t, committed) }).Return(cacheErr).Once()
					cache.On("DeleteManyWithContext", ctx, []string{rediskeys.PermissionsByClientID("client")}).Return(cacheErr).Once()
					var e error
					if stage == "response error" {
						e = failure
					}
					permissions.On("FindByClientID", ctx, "client").Return([]interfaces.PermissionResource{{ID: "actual", Name: "Actual permission", ResourceName: "Records"}}, e).Once()
				}
			}
			s := NewClientService(&configs.DefaultConfig{}, reg, cache)
			result, err := s.UpdateClient(ctx, req)
			switch stage {
			case "custom", "default", "empty", "nil active", "cache error":
				require.NoError(t, err)
				require.Equal(t, "Updated", result.Name)
				require.Equal(t, []models.ClientPermission{{ID: "actual", PermissionName: "Actual permission", ResourceName: "Records"}}, result.Permissions)
				require.Equal(t, stage != "nil active", result.IsActive)
			default:
				require.Error(t, err)
				require.Nil(t, result)
				if stage == "missing client" || stage == "zero update" {
					require.ErrorIs(t, err, domainerrors.ErrClientNotFound)
				}
			}
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
