package user

import (
	"context"
	"fmt"

	"github.com/ggwhite/go-masker/v2"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/shomali11/util/xhashes"
)

func (s *userService) RenderVerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) (*models.VerifyEmailResult, error) {
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.EmailVerifyPrefix, tokenHash)
	tokenData := models.EmailVerifyTokenData{}
	found, err := s.redisService.GetData(ctx, redisKey, &tokenData)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to get verify email token from redis", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if !found {
		log.WithContext(ctx).Error("Verify email token not found")
		return nil, errors.ErrInvalidVerifyEmailToken
	}
	log.WithContext(ctx).Debugw("Verify email token found", "key", redisKey, "user_id", tokenData.UserID)
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, tokenData.UserID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID", "error", err, "user_id", tokenData.UserID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", tokenData.UserID)
		return nil, errors.ErrInvalidVerifyEmailToken
	}
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, user.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err, "project_id", user.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", user.ProjectID)
		return nil, errors.ErrInvalidVerifyEmailToken
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "project_id", user.ProjectID)
		return nil, errors.ErrInvalidVerifyEmailToken
	}
	emailMasker := &masker.EmailMasker{}
	maskedEmail := emailMasker.Marshal("*", *user.Email)
	result := &models.VerifyEmailResult{
		Email:    maskedEmail,
		Project:  project,
		LoginURL: tokenData.LoginURL,
	}
	// If already verified, no need to update
	if user.EmailVerifiedAt != nil {
		err = s.redisService.DeleteWithContext(ctx, redisKey)
		if err != nil {
			log.WithContext(ctx).Warnw("Failed to delete verify email token from redis", "error", err, "redis_key", redisKey)
		}
		return result, nil
	}
	// Mark user's email as verified
	affected, err := userRepo.SetEmailVerified(ctx, user.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to set user as verified", "error", err, "user_id", user.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("Failed to set user as verified, no rows affected", "user_id", user.ID)
		return nil, errors.ErrInvalidVerifyEmailToken
	}

	// Invalidate user cache after successful email verification
	shared.InvalidateUserCache(ctx, s.redisService, user)

	// Delete the token from redis
	// Failed to delete token should not block the process
	err = s.redisService.DeleteWithContext(ctx, redisKey)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to delete verify email token from redis", "error", err, "redis_key", redisKey)
	}
	return result, nil
}
