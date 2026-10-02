package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetClients(t *testing.T) {
	for _, stage := range []string{"list", "empty", "count error", "list error", "list no account"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("offline")
			reg := rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			clients := im.NewMockClientRepository(t)
			desc := "Application"
			stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			stored := &entities.Client{ID: "client", ProjectID: "project", Name: "App", Description: &desc, IsActive: true, IsDefault: true, CreatedAt: stamp, UpdatedAt: stamp}
			req := &models.GetClientsRequest{ProjectID: "project", Search: "App"}
			noAccount := stage == "list no account"
			if noAccount {
				ctx = context.Background()
			} else {
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("ClientRepository").Return(clients).Once()
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
			}
			s := NewClientService(&configs.DefaultConfig{}, reg, nil)
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
		})
	}
}
