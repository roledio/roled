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
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	redism "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteClientProjectGuards(t *testing.T) {
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
			err := s.DeleteClient(ctx, &models.DeleteClientRequest{ProjectID: "project", ClientID: "client"})
			require.ErrorIs(t, err, expected)
		})
	}
}

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
