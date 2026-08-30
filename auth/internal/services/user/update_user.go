package user

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
)

func (s *userService) UpdateUser(ctx context.Context, req *models.UpdateUserRequest) (*models.UserDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// System project users cannot be updated
		log.WithContext(ctx).Error("Updating system project users are not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByIDAndProjectID(ctx, req.UserID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found by ID and project ID",
			"user_id", req.UserID,
			"project_id", project.ID)
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
	var ptrExternalUserID *string
	var ptrPasswordHash = user.PasswordHash

	// Validate email if provided
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email != "" {
		ptrEmail = &email

		// If email changed, check uniqueness
		if user.Email == nil || *user.Email != email {
			ptrEmailVerifiedAt = nil
			existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, project.ID, email)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find user by project ID and email", "error", err)
				return nil, pkgerrors.ErrSystemError.WithError(err)
			}
			if existingUser != nil {
				log.WithContext(ctx).Errorw("User with the same email already exists in the project",
					"project_id", project.ID,
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

	// Validate external user id if provided
	externalID := strings.TrimSpace(req.ExternalUserID)
	if externalID != "" {
		ptrExternalUserID = &externalID

		// If external_user_id changed, check uniqueness
		if user.ExternalUserID == nil || *user.ExternalUserID != externalID {
			existingUser, err := userRepo.FindByProjectIDAndExternalUserID(ctx, project.ID, externalID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find user by project ID and external user ID", "error", err)
				return nil, pkgerrors.ErrSystemError.WithError(err)
			}
			if existingUser != nil {
				log.WithContext(ctx).Errorw("User with the same external user ID already exists in the project",
					"project_id", project.ID, "external_user_id", externalID)
				return nil, errors.ErrUserExternalIDAlreadyUsed
			}
		}
	}

	// Validate role if provided
	var role *entities.Role
	if req.RoleID != "" {
		roleRepo := s.registry.RoleRepository()
		role, err = roleRepo.FindByIDAndProjectID(ctx, req.RoleID, project.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find role by ID and project ID", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if role == nil {
			log.WithContext(ctx).Errorw("Role not found for given ID and project ID",
				"role_id", req.RoleID, "project_id", project.ID)
			return nil, errors.ErrRoleNotFound
		}
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Update user entity
		user.Email = ptrEmail
		user.EmailVerifiedAt = ptrEmailVerifiedAt
		user.ExternalUserID = ptrExternalUserID
		user.PasswordHash = ptrPasswordHash
		user.DisplayName = strings.TrimSpace(req.DisplayName)
		user.AvatarURL = newAvatarURL
		user.IsActive = *req.IsActive

		affected, err := registry.UserRepository().Update(ctx, user)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update user", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			return errors.ErrUserNotFound
		}

		// Update role association
		userRoleRepo := registry.UserRoleRepository()
		_, err = userRoleRepo.DeleteByUserID(ctx, user.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete existing user role", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		if req.RoleID != "" {
			userRole := &entities.UserRole{
				UserID: user.ID,
				RoleID: req.RoleID,
			}
			if err := userRoleRepo.Create(ctx, userRole); err != nil {
				log.WithContext(ctx).Errorw("Failed to assign new role to user", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
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
	if role != nil {
		res.RoleID = role.ID
		res.RoleName = role.Name
	}

	return &res, nil
}

func (s *userService) deleteFile(ctx context.Context, avatarURL string) error {
	// The avatar URL is expected to be in the format of {uploadBaseURL}/{filePath}, so we need to trim the uploadBaseURL to get the file path
	filePath := strings.TrimPrefix(avatarURL, s.uploadBaseURL+"/")
	if filePath == avatarURL { // The file path is the same as the avatar URL, unable to trim the upload base URL prefix
		log.WithContext(ctx).Debugw("Avatar URL does not contain upload base URL prefix", "avatar_url", avatarURL, "upload_base_url", s.uploadBaseURL)
		return nil // The avatar URL does not contain the expected prefix, so we cannot determine the file path to delete. Log the error and return nil to avoid blocking the main flow.
	}
	err := s.uploadService.Delete(ctx, filePath)
	if err != nil {
		return err
	}
	log.WithContext(ctx).Debugw("File deleted successfully", "file_path", filePath)
	return nil
}
