package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	redism "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteClientTransaction(t *testing.T) {
	for _, stage := range []string{"success", "lookup error", "missing", "default", "tokens", "permissions", "client", "zero rows", "commit"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("unavailable")
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			clients := im.NewMockClientRepository(t)
			txClients := im.NewMockClientRepository(t)
			tokens := im.NewMockAccessTokenRepository(t)
			links := im.NewMockClientPermissionRepository(t)
			cache := redism.NewMockService(t)
			reg.On("ProjectRepository").Return(projects).Once()
			projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
			reg.On("ClientRepository").Return(clients).Once()
			stored := &entities.Client{ID: "client", ProjectID: "project", IsDefault: stage == "default"}
			if stage == "missing" {
				stored = nil
			}
			var lookupErr error
			if stage == "lookup error" {
				lookupErr = failure
			}
			clients.On("FindByIDAndProjectID", ctx, "client", "project").Return(stored, lookupErr).Once()
			committed := false
			if stage != "lookup error" && stage != "missing" && stage != "default" {
				tx.On("AccessTokenRepository").Return(tokens).Once()
				var e error
				if stage == "tokens" {
					e = failure
				}
				tokens.On("DeleteByClientID", ctx, "client").Return(2, e).Once()
				if stage != "tokens" {
					tx.On("ClientPermissionRepository").Return(links).Once()
					e = nil
					if stage == "permissions" {
						e = failure
					}
					links.On("DeleteByClientID", ctx, "client").Return(2, e).Once()
					if stage != "permissions" {
						tx.On("ClientRepository").Return(txClients).Once()
						e = nil
						if stage == "client" {
							e = failure
						}
						n := 1
						if stage == "zero rows" {
							n = 0
						}
						txClients.On("Delete", ctx, stored).Return(n, e).Once()
					}
				}
				reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
					if err := fn(tx); err != nil {
						return err
					}
					if stage == "commit" {
						return failure
					}
					committed = true
					return nil
				}).Once()
				if stage == "success" {
					cache.On("DeleteManyWithContext", ctx, []string{rediskeys.ClientByID("client"), rediskeys.ClientByIDAndProjectID("client", "project")}).Run(func(mock.Arguments) { require.True(t, committed) }).Return(nil).Once()
					cache.On("DeleteManyWithContext", ctx, []string{rediskeys.PermissionsByClientID("client")}).Return(nil).Once()
				}
			}
			err := NewClientService(&configs.DefaultConfig{}, reg, cache).DeleteClient(ctx, &models.DeleteClientRequest{ProjectID: "project", ClientID: "client"})
			switch stage {
			case "success":
				require.NoError(t, err)
			case "missing", "zero rows":
				require.ErrorIs(t, err, domainerrors.ErrClientNotFound)
			case "default":
				require.ErrorIs(t, err, domainerrors.ErrDeleteDefaultClient)
			case "commit":
				require.ErrorIs(t, err, failure)
			default:
				require.ErrorIs(t, err, pkgerrors.ErrSystemError)
			}
		})
	}
}

func TestClientListAndDetails(t *testing.T) {
	for _, stage := range []string{"list", "empty", "count error", "list error", "details", "missing", "lookup error", "invalid secret", "permissions error", "list no account", "details no account"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("offline")
			reg := rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			clients := im.NewMockClientRepository(t)
			links := im.NewMockClientPermissionRepository(t)
			master := "12345678901234567890123456789012"
			key, err := encryptionutil.DeriveKey([]byte(master), pkgconstants.KeyPurposeClientSecret)
			require.NoError(t, err)
			secret, err := encryptionutil.EncryptAES("private-secret", key, pkgconstants.KeyPurposeClientSecret)
			require.NoError(t, err)
			desc := "Application"
			stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			stored := &entities.Client{ID: "client", ProjectID: "project", Name: "App", Description: &desc, IsActive: true, IsDefault: true, SecretEncrypted: secret, CreatedAt: stamp, UpdatedAt: stamp}
			req := &models.GetClientsRequest{ProjectID: "project", Search: "App"}
			listMode := stage == "list" || stage == "empty" || stage == "count error" || stage == "list error" || stage == "list no account"
			noAccount := stage == "list no account" || stage == "details no account"
			if noAccount {
				ctx = context.Background()
			} else {
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("ClientRepository").Return(clients).Once()
				if listMode {
					n := 1
					var e error
					if stage == "empty" {
						n = 0
					}
					if stage == "count error" {
						e = failure
					}
					clients.On("Count", ctx, req).Return(n, e).Once()
					if n > 0 && e == nil {
						if stage == "list error" {
							e = failure
						}
						clients.On("FindAll", ctx, req).Return([]entities.Client{*stored}, e).Once()
					}
				} else {
					if stage == "missing" {
						stored = nil
					}
					if stage == "invalid secret" {
						stored.SecretEncrypted = "corrupt"
					}
					var e error
					if stage == "lookup error" {
						e = failure
					}
					clients.On("FindByIDAndProjectID", ctx, "client", "project").Return(stored, e).Once()
					if stage == "details" || stage == "permissions error" {
						reg.On("ClientPermissionRepository").Return(links).Once()
						if stage == "permissions error" {
							e = failure
						}
						links.On("FindByClientID", ctx, "client").Return([]entities.ClientPermission{{ClientID: "client", PermissionID: "read"}}, e).Once()
					}
				}
			}
			s := NewClientService(&configs.DefaultConfig{EncryptionMasterKey: master}, reg, nil)
			if listMode {
				got, total, e := s.GetClients(ctx, req)
				switch stage {
				case "list":
					require.NoError(t, e)
					require.Equal(t, 1, total)
					require.Equal(t, []models.ClientDetails{{ID: "client", Name: "App", Description: &desc, IsActive: true, IsDefault: true, CreatedAt: stamp, UpdatedAt: stamp}}, got)
				case "empty":
					require.NoError(t, e)
					require.Zero(t, total)
					require.Empty(t, got)
				default:
					require.Error(t, e)
					require.Nil(t, got)
					require.Zero(t, total)
				}
			} else {
				got, e := s.GetClientDetails(ctx, &models.GetClientDetailsRequest{ProjectID: "project", ClientID: "client"})
				if stage == "details" {
					require.NoError(t, e)
					require.Equal(t, &models.ClientDetails{ID: "client", Name: "App", Description: &desc, IsActive: true, IsDefault: true, CreatedAt: stamp, UpdatedAt: stamp, Secret: "private-secret", PermissionIDs: []string{"read"}}, got)
				} else {
					require.Error(t, e)
					require.Nil(t, got)
				}
			}
		})
	}
}
