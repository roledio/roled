package accesstoken

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgconstants "github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
)

func (s *accessTokenService) handleClientCredentials(ctx context.Context, req *models.ExchangeTokenRequest, client *entities.Client,
	project *entities.Project) (*models.ExchangeTokenResponse, error) {
	key, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), pkgconstants.KeyPurposeClientSecret)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to derive key for client secret decryption", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	clientSecret, err := encryptionutil.DecryptAES(client.SecretEncrypted, key, pkgconstants.KeyPurposeClientSecret)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to decrypt client secret", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if req.ClientSecret != clientSecret {
		log.WithContext(ctx).Errorw("Invalid client secret", "client_id", req.ClientID)
		return nil, errors.ErrInvalidClientCredentials
	}
	// Build access token with account ID from the client
	accessToken := s.buildAccessToken(ctx, req.GrantType, client, project, client.AccountID, nil)
	accessTokenRepo := s.registry.AccessTokenRepository()
	err = accessTokenRepo.Create(ctx, accessToken)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to create new access token", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	jwt, err := s.buildJWT(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return &models.ExchangeTokenResponse{
		AccessToken: jwt,
		TokenType:   tokenTypeBearer,
		ExpiresIn:   *accessToken.ExpiresIn,
	}, nil
}
