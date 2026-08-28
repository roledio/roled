package accesstoken

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/shomali11/util/xhashes"
)

func (s *accessTokenService) RevokeCurrentToken(ctx context.Context, req *models.RevokeCurrentTokenRequest) error {
	client, err := s.registry.ClientRepository().FindByID(ctx, req.ClientID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by ID", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Warnw("Client not found by ID", "client_id", req.ClientID)
		return nil // Return success for idempotent sign out
	}
	refreshTokenHash := xhashes.SHA256(req.RefreshToken)
	refreshToken, err := s.registry.RefreshTokenRepository().FindByClientIDAndRefreshTokenHash(ctx, req.ClientID, refreshTokenHash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find refresh token by client ID and refresh token hash", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if refreshToken == nil {
		log.WithContext(ctx).Warn("Refresh token not found by client ID and refresh token hash")
		return nil // Return success for idempotent sign out
	}
	if req.JWTClaims != nil {
		if refreshToken.ClientID != req.JWTClaims.ClientID {
			log.WithContext(ctx).Warnw("Refresh token client ID does not match JWT claims client ID",
				"refresh_token_client_id", refreshToken.ClientID,
				"jwt_client_id", req.JWTClaims.ClientID)
			return nil // Return success for idempotent sign out
		}
		if refreshToken.UserID != nil && req.JWTClaims.UserID != "" && *refreshToken.UserID != req.JWTClaims.UserID {
			log.WithContext(ctx).Warnw("Refresh token user ID does not match JWT claims user ID",
				"refresh_token_user_id", refreshToken.UserID,
				"jwt_user_id", req.JWTClaims.UserID)
			return nil // Return success for idempotent sign out
		}
	}
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Revoke the refresh token
		_, err := registry.RefreshTokenRepository().UpdateAsRevoked(ctx, refreshToken)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to revoke refresh token", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Revoke the access token
		if req.JWTClaims != nil {
			_, err = registry.AccessTokenRepository().UpdateAsRevoked(ctx, req.JWTClaims.ID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to revoke access token", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Invalidate caches after successful revocation
	shared.InvalidateRefreshTokenCache(ctx, s.redisService, req.ClientID, refreshTokenHash)
	if req.JWTClaims != nil {
		shared.InvalidateAccessTokenCache(ctx, s.redisService, req.JWTClaims.ID)
	}

	return nil
}
