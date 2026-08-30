package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/shomali11/util/xhashes"
)

func (s *userService) RenderResetPassword(ctx context.Context, req *models.RenderResetPasswordRequest) (*models.RenderResetPasswordResult, error) {
	if req.ProjectID != nil {
		// If the project ID is provided, skip token validation
		// This is used when user has successfully reset their password (token is already deleted from redis)
		// and is redirected to the reset password page again
		projectRepo := s.registry.ProjectRepository()
		project, err := projectRepo.FindByID(ctx, *req.ProjectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err, "project_id", *req.ProjectID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if project == nil {
			log.WithContext(ctx).Errorw("Project not found", "project_id", *req.ProjectID)
			return nil, errors.ErrProjectNotFound
		}
		return &models.RenderResetPasswordResult{
			Project: project,
		}, nil
	}
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ResetPasswordPrefix, tokenHash)
	_, _, project, err := s.validateResetPassword(ctx, redisKey)
	if err != nil {
		return nil, err
	}
	result := &models.RenderResetPasswordResult{
		Project: project,
	}
	return result, nil
}

func (s *userService) validateResetPassword(ctx context.Context, redisKey string) (*models.ResetPasswordTokenData, *entities.User, *entities.Project, error) {
	tokenData := models.ResetPasswordTokenData{}
	found, err := s.redisService.GetData(ctx, redisKey, &tokenData)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to get reset password token from redis", "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if !found {
		log.WithContext(ctx).Error("Reset password token not found")
		return nil, nil, nil, errors.ErrInvalidResetPasswordToken
	}
	uid := tokenData.UserID
	log.WithContext(ctx).Debugw("Reset password token found", "key", redisKey, "user_id", uid)

	// Validate user
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, uid)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID", "error", err, "user_id", uid)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", uid)
		return nil, nil, nil, errors.ErrUserNotFound
	}
	if !user.IsActive {
		log.WithContext(ctx).Errorw("User is not active", "user_id", uid)
		return nil, nil, nil, errors.ErrUserNotActive
	}

	// Validate account
	accountRepo := s.registry.AccountRepository()
	account, err := accountRepo.FindByID(ctx, user.AccountID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find account by ID", "error", err, "account_id", user.AccountID)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if account == nil {
		log.WithContext(ctx).Errorw("Account not found for user", "user_id", uid, "account_id", user.AccountID)
		return nil, nil, nil, errors.ErrAccountNotFound
	}
	if !account.IsActive {
		log.WithContext(ctx).Errorw("Account is not active for user", "user_id", uid, "account_id", user.AccountID)
		return nil, nil, nil, errors.ErrAccountNotActive
	}

	// Validate project
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, user.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err, "project_id", user.ProjectID)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", user.ProjectID)
		return nil, nil, nil, errors.ErrProjectNotFound
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "project_id", user.ProjectID)
		return nil, nil, nil, errors.ErrProjectNotActive
	}

	return &tokenData, user, project, nil
}
