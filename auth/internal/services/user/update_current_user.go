package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
)

func (s *userService) UpdateCurrentUser(ctx context.Context, req *models.UpdateCurrentUserRequest) (*models.UserDetails, error) {
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		return nil, errors.ErrCtxAccessTokenNotFound
	}
	if accessToken.UserID == nil {
		err := fmt.Errorf("non-user access token is not supported for getting current user details")
		log.WithContext(ctx).Errorw("Failed to get current user details", "error", err)
		return nil, pkgerrors.ErrOperationNotAvailable.WithError(err)
	}
	userID := *accessToken.UserID
	projectID := accessToken.ProjectID
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByIDAndProjectID(ctx, userID, projectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found by ID and project ID",
			"user_id", userID,
			"project_id", projectID)
		return nil, errors.ErrUserNotFound
	}

	var tmpAvatarURL string
	if req.AvatarURL != "" {
		tmpAvatarURL = req.AvatarURL
	}
	newAvatarURL, isTmpAvatarURL := s.checkUploadAvatarURL(req.AvatarURL)
	oldAvatarURL := user.AvatarURL

	var ptrEmailVerifiedAt = user.EmailVerifiedAt
	var ptrEmail *string
	var ptrPasswordHash = user.PasswordHash

	// Validate email if provided
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email != "" {
		ptrEmail = &email

		// If email changed, check uniqueness
		if user.Email == nil || *user.Email != email {
			ptrEmailVerifiedAt = nil
			existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, projectID, email)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find user by project ID and email", "error", err)
				return nil, pkgerrors.ErrSystemError.WithError(err)
			}
			if existingUser != nil {
				log.WithContext(ctx).Errorw("User with the same email already exists in the project",
					"project_id", projectID,
					"email", email)
				return nil, errors.ErrUserEmailAlreadyUsed
			}
		}
	}

	// Update password if provided
	if req.Password != "" {
		hash, err := passwordutil.HashPassword(req.Password)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to hash password", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		ptrPasswordHash = &hash
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Update user entity
		user.Email = ptrEmail
		user.EmailVerifiedAt = ptrEmailVerifiedAt
		user.PasswordHash = ptrPasswordHash
		user.DisplayName = strings.TrimSpace(req.DisplayName)
		user.AvatarURL = newAvatarURL

		affected, err := registry.UserRepository().Update(ctx, user)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update user", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			return errors.ErrUserNotFound
		}

		// Make sure to move the avatar file after the user is successfully updated
		// If it is already moved and the transaction fails, the file cannot be used
		// again using the same avatar URL
		if isTmpAvatarURL && tmpAvatarURL != "" {
			err = s.moveFileFromTmp(ctx, tmpAvatarURL)
			if err != nil {
				// The error should rollback the transaction since the file must be successfully moved
				// or the avatar URL will not be able to be accessed using the new URL
				return err
			}
		}

		// If the avatar URL is changed (or become empty) and the old avatar URL is not nil, delete the old avatar file
		if oldAvatarURL != nil && (newAvatarURL == nil || *oldAvatarURL != *newAvatarURL) {
			err = s.deleteFile(ctx, *oldAvatarURL)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to delete old avatar file", "error", err, "avatar_url", *oldAvatarURL)
				// The error should not rollback the transaction since the user has been successfully updated, the old avatar file can be deleted later
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cache after successful update
	shared.InvalidateUserCache(ctx, s.redisService, user)

	res := models.UserDetails{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       time.Now().UTC(),
		Email:           user.Email,
		ExternalUserID:  user.ExternalUserID,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		IsActive:        user.IsActive,
		IsEmailVerified: user.EmailVerifiedAt != nil,
	}

	return &res, nil
}
