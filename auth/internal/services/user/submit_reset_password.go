package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
	"github.com/shomali11/util/xhashes"
)

func (s *userService) SubmitResetPassword(ctx context.Context, req *models.SubmitResetPasswordRequest) (*models.SubmitResetPasswordResult, error) {
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ResetPasswordPrefix, tokenHash)
	tokenData, user, project, err := s.validateResetPassword(ctx, redisKey)
	if err != nil {
		return nil, err
	}
	// Update user's password
	passwordHash, err := passwordutil.HashPassword(req.Password)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to hash new password", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	userRepo := s.registry.UserRepository()
	affected, err := userRepo.UpdatePassword(ctx, user.ID, passwordHash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update user password", "error", err, "user_id", user.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("Failed to update user password, no rows affected", "user_id", user.ID)
		return nil, errors.ErrInvalidResetPasswordToken
	}

	// Invalidate user cache after successful password update
	shared.InvalidateUserCache(ctx, s.redisService, user, &shared.OldUserCacheKeyParts{
		Email:          user.Email,
		ExternalUserID: user.ExternalUserID,
	})

	// Delete the token from redis
	// Failed to delete token should not block the process
	err = s.redisService.DeleteWithContext(ctx, redisKey)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to delete reset password token from redis", "error", err, "redis_key", redisKey)
	} else {
		log.WithContext(ctx).Debugw("Successfully deleted reset password token from redis", "redis_key", redisKey)
	}
	result := models.SubmitResetPasswordResult{
		Project:  project,
		LoginURL: tokenData.LoginURL,
	}
	return &result, nil
}
