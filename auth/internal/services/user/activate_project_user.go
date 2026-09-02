package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
	"github.com/shomali11/util/xhashes"
)

// RenderActivateProjectUser retrieves user and project data for rendering the activation form
func (s *userService) RenderActivateProjectUser(ctx context.Context, req *models.RenderActivateProjectUserRequest) (*models.RenderActivateProjectUserResponse, error) {

	// Get token data from Redis
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateProjectUserPrefix, tokenHash)
	var tokenData models.ActivateProjectUserTokenData
	found, err := s.redisService.GetData(ctx, redisKey, &tokenData)
	if err != nil || !found {
		log.WithContext(ctx).Errorw("Failed to retrieve activation token from redis", "error", err, "token_hash", tokenHash)
		return nil, errors.ErrInvalidActivationToken
	}

	// If UserID is provided in request, use it; otherwise use the one from token
	userID := tokenData.UserID
	if req.UserID != nil && *req.UserID != "" {
		userID = *req.UserID
	}

	// Find user
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user", "error", err, "user_id", userID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", userID)
		return nil, errors.ErrUserNotFound
	}

	// Find project
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, user.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project", "error", err, "project_id", user.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", user.ProjectID)
		return nil, errors.ErrProjectNotFound
	}

	response := &models.RenderActivateProjectUserResponse{
		User:     user,
		Project:  project,
		LoginURL: tokenData.LoginURL,
	}

	return response, nil
}

// SubmitActivateProjectUser activates the user by setting password and marking user as active
func (s *userService) SubmitActivateProjectUser(ctx context.Context, req *models.SubmitActivateProjectUserRequest) (*models.SubmitActivateProjectUserResponse, error) {

	// Validate token and get token data
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateProjectUserPrefix, tokenHash)
	var tokenData models.ActivateProjectUserTokenData
	found, err := s.redisService.GetData(ctx, redisKey, &tokenData)
	if err != nil || !found {
		log.WithContext(ctx).Errorw("Failed to retrieve activation token from redis", "error", err, "token_hash", tokenHash)
		return nil, errors.ErrInvalidActivationToken
	}

	userID := tokenData.UserID

	// Find user
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user", "error", err, "user_id", userID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", userID)
		return nil, errors.ErrUserNotFound
	}

	// Check if user is already active (password already set)
	if user.IsActive && user.PasswordHash != nil {
		log.WithContext(ctx).Errorw("User is already activated", "user_id", userID)
		return nil, errors.ErrUserAlreadyActive
	}

	// Hash the new password
	passwordHash, err := passwordutil.HashPassword(req.Password)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to hash password", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Update user with password and mark as active
	now := time.Now()
	user.PasswordHash = &passwordHash
	user.DisplayName = req.DisplayName
	user.IsActive = true
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now

	err = s.registry.Tx(func(registry repositories.Registry) error {
		_, err := registry.UserRepository().Update(ctx, user)
		return err
	})

	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update user", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Delete the token from Redis
	err = s.redisService.DeleteWithContext(ctx, redisKey)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to delete activation token from redis", "error", err)
		// Don't return error here, token deletion failure should not block activation
	}

	// Find project for response
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, user.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project", "error", err, "project_id", user.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", user.ProjectID)
		return nil, errors.ErrProjectNotFound
	}

	response := &models.SubmitActivateProjectUserResponse{
		UserID:   userID,
		Project:  project,
		LoginURL: tokenData.LoginURL,
	}

	return response, nil
}
