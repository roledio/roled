package accesstoken

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/pkceutil"
	"github.com/shomali11/util/xhashes"
)

func (s *accessTokenService) handleAuthorizationCode(ctx context.Context, req *models.ExchangeTokenRequest, client *entities.Client,
	project *entities.Project) (*models.ExchangeTokenResponse, error) {
	authCodeRepo := s.registry.AuthCodeRepository()
	codeHash := xhashes.SHA256(req.AuthorizationCode)
	authCode, err := authCodeRepo.FindByClientIDAndCodeHash(ctx, client.ID, codeHash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find auth code by client ID and code hash", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if authCode == nil {
		log.WithContext(ctx).Errorw("Authorization code not found", "code", req.AuthorizationCode)
		return nil, errors.ErrInvalidAuthorizationCode
	}
	if authCode.ClientID != req.ClientID {
		log.WithContext(ctx).Errorw("Client ID mismatch", "expected", authCode.ClientID, "got", req.ClientID)
		return nil, errors.ErrInvalidAuthorizationCode
	}
	if authCode.RedirectURI != req.RedirectURI {
		log.WithContext(ctx).Errorw("Redirect URI mismatch", "expected", authCode.RedirectURI, "got", req.RedirectURI)
		return nil, errors.ErrDifferentAuthorizeRedirectURI
	}
	now := time.Now()
	if authCode.ExpiresAt.Before(now) {
		log.WithContext(ctx).Errorw("Authorization code is expired", "code", req.AuthorizationCode)
		return nil, errors.ErrAuthCodeExpired
	}
	if authCode.UsedAt != nil {
		log.WithContext(ctx).Errorw("Authorization code has been used", "code", req.AuthorizationCode)
		return nil, errors.ErrAuthCodeAlreadyUsed
	}
	// Verify PKCE
	if !pkceutil.VerifyCodeChallenge(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
		log.WithContext(ctx).Errorw("Code challenge verification failed", "code_verifier", req.CodeVerifier)
		return nil, errors.ErrInvalidCodeVerifier
	}

	var result *models.ExchangeTokenResponse
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Create a new auth code repository instance within the transaction
		authCodeRepo := registry.AuthCodeRepository()
		// Mark auth code as used
		affected, err := authCodeRepo.UpdateUsedAuthCode(ctx, authCode)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to mark auth code as used", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("Failed to mark auth code as used, no rows affected", "auth_code_id", authCode.ID)
			return errors.ErrInvalidAuthorizationCode
		}
		// Create a new refresh token
		newRefreshToken, refreshTokenPlain := s.buildRefreshToken(ctx, client, project, authCode.AccountID, authCode.UserID)
		refreshTokenRepo := registry.RefreshTokenRepository()
		err = refreshTokenRepo.Create(ctx, newRefreshToken)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create new refresh token", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Create a new access token with account ID from the auth code (user's account id)
		accessTokenRepo := registry.AccessTokenRepository()
		newAccessToken := s.buildAccessToken(ctx, req.GrantType, client, project, authCode.AccountID, authCode.UserID)
		newAccessToken.RefreshTokenID = &newRefreshToken.ID
		newAccessToken.AuthCodeID = &authCode.ID
		err = accessTokenRepo.Create(ctx, newAccessToken)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create new access token", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Build JWT
		jwt, err := s.buildJWT(ctx, newAccessToken)
		if err != nil {
			return err
		}
		result = &models.ExchangeTokenResponse{
			AccessToken:           jwt,
			TokenType:             tokenTypeBearer,
			ExpiresIn:             *newAccessToken.ExpiresIn,
			RefreshToken:          refreshTokenPlain,
			RefreshTokenExpiresIn: *newRefreshToken.ExpiresIn,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate the used auth code cache after successful exchange
	shared.InvalidateAuthCodeCache(ctx, s.redisService, client.ID, codeHash)

	return result, nil
}
