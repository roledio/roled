package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
)

func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserDetails, error) {
	account, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// System project users are created automatically when creating member.
		// Creating user for system project is not possible.
		log.WithContext(ctx).Errorw("Creating a new user for system project is not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	var tmpAvatarURL string
	if req.AvatarURL != "" {
		tmpAvatarURL = req.AvatarURL
	}
	newAvatarURL, isTmp := s.checkUploadAvatarURL(req.AvatarURL)

	var ptrExternalUserID *string
	var ptrEmail *string
	var ptrPasswordHash *string

	userRepo := s.registry.UserRepository()
	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))

		// Check duplicate email in this project
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

		ptrEmail = &email
	}

	// Hash password if provided
	if req.Password != "" {
		hash, err := passwordutil.HashPassword(req.Password)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to hash password", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		ptrPasswordHash = &hash
	}

	externalID := strings.TrimSpace(req.ExternalUserID)
	if externalID != "" {
		// Check duplicate external_user_id in this project
		existingUser, err := userRepo.FindByProjectIDAndExternalUserID(ctx, project.ID, externalID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find user by project ID and external user ID", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if existingUser != nil {
			log.WithContext(ctx).Errorw("User with the same external_user_id already exists in the project",
				"project_id", project.ID, "external_user_id", externalID)
			return nil, errors.ErrUserExternalIDAlreadyUsed
		}

		ptrExternalUserID = &externalID
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

	var newUser *entities.User
	err = s.registry.Tx(func(registry repositories.Registry) error {
		newUser = &entities.User{
			ID:             idutil.NewID(),
			AccountID:      account.ID,
			ProjectID:      project.ID,
			Email:          ptrEmail,
			PasswordHash:   ptrPasswordHash,
			ExternalUserID: ptrExternalUserID,
			DisplayName:    strings.TrimSpace(req.DisplayName),
			AvatarURL:      newAvatarURL,
			IsActive:       true,
		}
		if err := registry.UserRepository().Create(ctx, newUser); err != nil {
			log.WithContext(ctx).Errorw("Failed to create user", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Assign role if provided
		if req.RoleID != "" {
			userRole := &entities.UserRole{
				UserID: newUser.ID,
				RoleID: req.RoleID,
			}
			if err := registry.UserRoleRepository().Create(ctx, userRole); err != nil {
				log.WithContext(ctx).Errorw("Failed to assign role to user", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}

		// Make sure to move the avatar file after the user is successfully created
		// If it is already moved and the transaction fails, the file cannot be used
		// again using the same avatar URL
		if isTmp && tmpAvatarURL != "" {
			err = s.moveFileFromTmp(ctx, tmpAvatarURL)
			if err != nil {
				// The error should rollback the transaction since the file must be successfully moved
				// or the avatar URL will not be able to be accessed using the new URL
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	res := models.UserDetails{
		ID:              newUser.ID,
		CreatedAt:       now,
		UpdatedAt:       now,
		Email:           newUser.Email,
		ExternalUserID:  newUser.ExternalUserID,
		DisplayName:     newUser.DisplayName,
		AvatarURL:       newUser.AvatarURL,
		IsActive:        newUser.IsActive,
		IsEmailVerified: false,
	}
	if role != nil {
		res.RoleID = role.ID
		res.RoleName = role.Name
	}
	return &res, nil
}

func (s *userService) checkUploadAvatarURL(avatarURL string) (*string, bool) {
	if avatarURL == "" {
		return nil, false
	}
	tmpUploadURL := s.uploadBaseURL + "/tmp"
	if after, ok := strings.CutPrefix(avatarURL, tmpUploadURL); ok {
		newURL := s.uploadBaseURL + after // Return new avatar URL without /tmp prefix
		return &newURL, true
	}
	return &avatarURL, false
}

func (s *userService) moveFileFromTmp(ctx context.Context, avatarURL string) error {
	// Extract the file path after /tmp
	tmpIndex := strings.Index(avatarURL, "tmp/")
	if tmpIndex == -1 {
		log.WithContext(ctx).Errorw("Invalid avatarURL: tmp/ not found", "avatarURL", avatarURL)
		return pkgerrors.ErrSystemError.WithError(fmt.Errorf("invalid avatarURL: tmp/ not found"))
	}
	tmpFilePath := avatarURL[tmpIndex:]
	newFilePath := strings.TrimPrefix(tmpFilePath, "tmp/")

	log.WithContext(ctx).Debugw("Moving file from tmp", "from", tmpFilePath, "to", newFilePath)

	// Move the file using upload service
	err := s.uploadService.Move(ctx, tmpFilePath, newFilePath)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to move file from tmp", "error", err, "from", tmpFilePath, "to", newFilePath)
		return errors.ErrMoveTmpUserAvatar.WithError(err)
	}
	log.WithContext(ctx).Debugw("File moved successfully from tmp", "from", tmpFilePath, "to", newFilePath)
	return nil
}
