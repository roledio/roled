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
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/stretchr/testify/require"
)

func TestGetClientDetails(t *testing.T) {
	for _, stage := range []string{"details", "missing", "lookup error", "invalid secret", "permissions error", "details no account"} {
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
			noAccount := stage == "details no account"
			if noAccount {
				ctx = context.Background()
			} else {
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("ClientRepository").Return(clients).Once()
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
			s := NewClientService(&configs.DefaultConfig{EncryptionMasterKey: master}, reg, nil)
			got, e := s.GetClientDetails(ctx, &models.GetClientDetailsRequest{ProjectID: "project", ClientID: "client"})
			if stage == "details" {
				require.NoError(t, e)
				require.Equal(t, &models.ClientDetails{ID: "client", Name: "App", Description: &desc, IsActive: true, IsDefault: true, CreatedAt: stamp, UpdatedAt: stamp, Secret: "private-secret", PermissionIDs: []string{"read"}}, got)
			} else {
				require.Error(t, e)
				require.Nil(t, got)
			}
		})
	}
}
