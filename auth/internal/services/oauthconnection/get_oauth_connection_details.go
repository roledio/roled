package oauthconnection

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
)

func (s *oAuthConnectionService) GetOAuthConnectionDetails(ctx context.Context, req *models.GetOAuthConnectionRequest) (*models.OAuthConnectionDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	oauthConnectionRepo := s.registry.OAuthConnectionRepository()
	connection, err := oauthConnectionRepo.FindByProjectIDAndProvider(ctx, req.ProjectID, req.Provider)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find OAuth connection", "error", err, "project_id", req.ProjectID, "provider", req.Provider)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if connection == nil {
		log.WithContext(ctx).Warnw("OAuth connection not found", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrOAuthConnectionNotFound
	}

	res := &models.OAuthConnectionDetails{
		ID:             connection.ID,
		CreatedAt:      connection.CreatedAt,
		UpdatedAt:      connection.UpdatedAt,
		ProjectID:      connection.ProjectID,
		Provider:       connection.Provider,
		CredentialType: connection.CredentialType,
		Enabled:        connection.Enabled,
	}

	if connection.CredentialType == constants.OAuthCredentialTypeCustom {

		// Decrypte oauth client secret
		purpose := constants.KeyPurposeOAuthClientSecret
		derivedKey, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), purpose)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to derive key for client secret encryption", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		clientSecret, err := encryptionutil.DecryptAES(*connection.ClientSecretEncrypted, derivedKey, purpose)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to decrypt client secret", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}

		res.ClientID = connection.ClientID
		res.ClientSecret = &clientSecret
		res.Scopes = strings.Fields(*connection.Scopes)
	}

	return res, nil
}
