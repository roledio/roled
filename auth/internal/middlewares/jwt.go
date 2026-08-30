package middlewares

import (
	"context"
	"strings"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
	"go.openly.dev/pointy"
)

func JWT(defaultConfig *configs.DefaultConfig, registry repositories.Registry) fiber.Handler {
	return jwtware.New(jwtware.Config{
		Next: func(c fiber.Ctx) bool {
			path := c.Path()
			return strings.HasPrefix(path, "/system") || path == "/api/v1/tokens"
		},
		SigningKey: jwtware.SigningKey{Key: []byte(defaultConfig.JWT.SigningKey)},
		Claims:     &models.JWTClaims{},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			log.WithContext(c.Context()).Errorw("Failed to validate JWT", "error", err)
			return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken.WithError(err))
		},
		SuccessHandler: func(c fiber.Ctx) error {
			ctx := c.Context()
			token := jwtware.FromContext(c)
			claims := token.Claims.(*models.JWTClaims)
			accessTokenRepo := registry.AccessTokenRepository()
			accessToken, err := accessTokenRepo.FindByID(ctx, claims.ID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find access token by ID", "error", err)
				return responseutil.SendError(c, errors.ErrSystemError.WithError(err))
			}
			if accessToken == nil {
				log.WithContext(ctx).Warnw("Access token not found", "id", claims.ID)
				return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
			}
			if accessToken.Status != constants.AccessTokenStatusIssued {
				log.WithContext(ctx).Warnw("Access token is not active", "id", claims.ID, "status", accessToken.Status)
				return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
			}
			// Validate that the token claims match the access token record
			if accessToken.ProjectID != claims.Audience[0] {
				log.WithContext(ctx).Warnw("Access token project ID does not match", "id", claims.ID, "expected_project_id", accessToken.ProjectID, "token_audience", claims.Audience)
				return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
			}
			if accessToken.ClientID != claims.ClientID {
				log.WithContext(ctx).Warnw("Access token client ID does not match", "id", claims.ID, "expected_client_id", accessToken.ClientID, "token_client_id", claims.ClientID)
				return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
			}
			if pointy.StringValue(accessToken.UserID, "") != claims.UserID {
				log.WithContext(ctx).Warnw("Access token user ID does not match", "id", claims.ID, "expected_user_id", accessToken.UserID, "token_user_id", claims.UserID)
				return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
			}
			accountRepo := registry.AccountRepository()
			account, err := accountRepo.FindByID(ctx, accessToken.AccountID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find account by ID", "error", err)
				return responseutil.SendError(c, errors.ErrSystemError.WithError(err))
			}
			if account == nil {
				log.WithContext(ctx).Errorw("Account not found for access token", "access_token_id", accessToken.ID, "account_id", accessToken.AccountID)
				return responseutil.SendError(c, errors.ErrSystemError)
			}
			ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)
			ctx = context.WithValue(ctx, constants.CtxAccount, account)
			c.SetContext(ctx)
			c.Locals(constants.CtxAccessToken, accessToken) // also set in locals for easier access in handlers
			return c.Next()
		},
	})
}
