package authorize

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	autherrors "github.com/roledio/roled/auth/internal/errors"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestGoogleOAuthConnectionConfiguration(t *testing.T) {
	for _, stage := range []string{"custom", "default", "lookup failure", "missing", "disabled", "custom secret invalid", "system lookup failure", "system missing", "system connection failure", "system connection missing", "system connection disabled", "system secret invalid"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			connections := im.NewMockOAuthConnectionRepository(t)
			projects := im.NewMockProjectRepository(t)
			reg.EXPECT().OAuthConnectionRepository().Return(connections)
			failure := errors.New("database unavailable")
			clientID, secret, scopes := "google-client", encryptClientSecret(t), "openid  email\tprofile"
			conn := &entities.OAuthConnection{ProjectID: "tenant", CredentialType: "custom", Enabled: true, ClientID: &clientID, ClientSecretEncrypted: &secret, Scopes: &scopes}
			var lookupErr error
			var expected error
			switch stage {
			case "lookup failure":
				lookupErr, expected = failure, pkgerrors.ErrSystemError
			case "missing":
				conn, expected = nil, autherrors.ErrProviderOAuthConnectionNotFound("google")
			case "disabled":
				conn.Enabled, expected = false, autherrors.ErrProviderOAuthConnectionDisabled("google")
			case "custom secret invalid":
				secret, expected = "corrupt ciphertext", pkgerrors.ErrSystemError
			case "custom":
			default:
				conn.CredentialType = "default"
				reg.EXPECT().ProjectRepository().Return(projects)
				var system = &entities.Project{ID: "system"}
				var systemErr error
				if stage == "system lookup failure" {
					systemErr, expected = failure, pkgerrors.ErrSystemError
				}
				if stage == "system missing" {
					system, expected = nil, pkgerrors.ErrSystemError
				}
				projects.EXPECT().FindSystem(ctx).Return(system, systemErr)
				if system != nil && systemErr == nil {
					clientID = "system-client"
					stored := &entities.OAuthConnection{ProjectID: "system", CredentialType: "custom", Enabled: true, ClientID: &clientID, ClientSecretEncrypted: &secret, Scopes: &scopes}
					var storedErr error
					switch stage {
					case "system connection failure":
						storedErr, expected = failure, pkgerrors.ErrSystemError
					case "system connection missing":
						stored, expected = nil, autherrors.ErrProviderOAuthConnectionNotFound("google")
					case "system connection disabled":
						stored.Enabled, expected = false, autherrors.ErrProviderOAuthConnectionDisabled("google")
					case "system secret invalid":
						secret, expected = "invalid", pkgerrors.ErrSystemError
					}
					connections.EXPECT().FindByProjectIDAndProvider(ctx, "system", "google").Return(stored, storedErr)
				}
			}
			connections.EXPECT().FindByProjectIDAndProvider(ctx, "tenant", "google").Return(conn, lookupErr)
			s := &authorizeService{registry: reg, defaultConfig: &configs.DefaultConfig{BaseURL: "https://auth.example", EncryptionMasterKey: "test-master-key-32-bytes-long!!"}}
			got, err := s.validateGoogleOAuthConnection(ctx, &entities.Project{ID: "tenant"})
			if expected != nil {
				require.ErrorIs(t, err, expected)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, clientID, got.ClientID)
			require.Equal(t, "google-client-secret", got.ClientSecret)
			require.Equal(t, []string{"openid", "email", "profile"}, got.Scopes)
			require.Equal(t, "https://auth.example/oauth/google/callback", got.RedirectURL)
		})
	}
}
