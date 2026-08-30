package accesstoken

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/shomali11/util/xhashes"
)

func (s *accessTokenService) handleRefreshToken(ctx context.Context, req *models.ExchangeTokenRequest, client *entities.Client,
	project *entities.Project) (*models.ExchangeTokenResponse, error) {
	refreshTokenRepo := s.registry.RefreshTokenRepository()
	refreshTokenHash := xhashes.SHA256(req.RefreshToken)
	refreshToken, err := refreshTokenRepo.FindByClientIDAndRefreshTokenHash(ctx, client.ID, refreshTokenHash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find refresh token by client ID and refresh token hash", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if refreshToken == nil {
		log.WithContext(ctx).Errorw("Refresh token not found by client ID and refresh token hash", "client_id", client.ID)
		return nil, errors.ErrInvalidRefreshToken
	}
	if refreshToken.Status != constants.RefreshTokenStatusIssued {
		log.WithContext(ctx).Errorw("Refresh token is not in issued status", "refresh_token_id", refreshToken.ID, "status", refreshToken.Status)
		return nil, errors.ErrInvalidRefreshToken
	}
	now := time.Now()
	expiredAt := refreshToken.IssuedAt.Add(time.Duration(*refreshToken.ExpiresIn) * time.Second)
	if expiredAt.Before(now) {
		log.WithContext(ctx).Errorw("Refresh token is expired", "refresh_token_id", refreshToken.ID)
		return nil, errors.ErrRefreshTokenExpired
	}
	if refreshToken.UsedAt != nil {
		log.WithContext(ctx).Errorw("Refresh token has been used", "refresh_token_id", refreshToken.ID)
		return nil, errors.ErrRefreshTokenAlreadyUsed
	}
	var result *models.ExchangeTokenResponse
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Create a new refresh token repository instance within the transaction
		refreshTokenRepo := registry.RefreshTokenRepository()
		// Mark current refresh token as used
		affected, err := refreshTokenRepo.UpdateUsedRefreshToken(ctx, refreshToken)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update refresh token as used", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("Failed to update refresh token as used, no rows affected", "refresh_token_id", refreshToken.ID)
			return errors.ErrInvalidRefreshToken
		}
		// Create a new refresh token
		newRefreshToken, refreshTokenPlain := s.buildRefreshToken(ctx, client, project, refreshToken.AccountID, refreshToken.UserID)
		err = refreshTokenRepo.Create(ctx, newRefreshToken)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create new refresh token", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Create a new access token with account ID from the refresh token (user's account id)
		accessTokenRepo := registry.AccessTokenRepository()
		newAccessToken := s.buildAccessToken(ctx, req.GrantType, client, project, refreshToken.AccountID, refreshToken.UserID)
		newAccessToken.RefreshTokenID = &newRefreshToken.ID
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

	// Invalidate the used refresh token cache after successful exchange
	shared.InvalidateRefreshTokenCache(ctx, s.redisService, client.ID, refreshTokenHash)

	return result, nil
}
