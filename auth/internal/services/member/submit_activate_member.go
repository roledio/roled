package member

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
	"github.com/shomali11/util/xhashes"
)

func (s *memberService) SubmitActivateMember(ctx context.Context, req *models.SubmitActivateMemberRequest) (*models.SubmitActivateMemberResponse, error) {
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateMemberPrefix, tokenHash)
	tokenData, err := s.readTokenData(ctx, redisKey)
	if err != nil {
		return nil, err
	}

	resp := models.SubmitActivateMemberResponse{
		UserID:   tokenData.UserID,
		LoginURL: tokenData.LoginURL,
	}

	result, err := s.prepareRenderActivateMember(ctx, tokenData.UserID)
	if err != nil {
		return nil, err
	}

	resp.Account = result.Account
	resp.Project = result.Project

	user := result.User
	if user.IsActive && user.EmailVerifiedAt != nil {
		log.WithContext(ctx).Debugw("User is already activated, update skipped", "user_id", user.ID)
		return &resp, nil
	}

	// Update user with password and display name
	passwordHash, err := passwordutil.HashPassword(req.Password)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to hash password", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	now := time.Now()
	user.DisplayName = req.DisplayName
	user.PasswordHash = &passwordHash
	user.IsActive = true
	user.EmailVerifiedAt = &now
	userRepo := s.registry.UserRepository()
	affected, err := userRepo.Update(ctx, user)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update user", "error", err, "user_id", user.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No user updated", "user_id", user.ID)
		return nil, errors.ErrUserNotFound
	}

	// Invalidate user cache after successful user update
	shared.InvalidateUserCache(ctx, s.redisService, user, &shared.OldUserCacheKeyParts{
		ExternalUserID: user.ExternalUserID,
		Email:          user.Email,
	})

	// Delete the token from redis
	// Failed to delete token should not block the process
	err = s.redisService.DeleteWithContext(ctx, redisKey)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to delete activate member token from redis", "error", err, "redis_key", redisKey)
	} else {
		log.WithContext(ctx).Debugw("Successfully deleted activate member token from redis", "redis_key", redisKey)
	}

	return &resp, nil
}
