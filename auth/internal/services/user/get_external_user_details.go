package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *userService) GetExternalUserDetails(ctx context.Context, req *models.GetExternalUserDetailsRequest) (*models.UserDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return s.processGetExternalUserDetails(ctx, req)
}

func (s *userService) processGetExternalUserDetails(ctx context.Context, req *models.GetExternalUserDetailsRequest) (*models.UserDetails, error) {
	user, err := s.registry.UserRepository().FindByProjectIDAndExternalUserIDJoinRole(ctx, req.ProjectID, req.ExternalUserID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by external user ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}
	var permissions []string
	if req.IncludePermissions && user.RoleID != "" {
		perms, err := s.registry.PermissionRepository().FindByRoleID(ctx, user.RoleID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find user permissions", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		permissions = make([]string, len(perms))
		for i, p := range perms {
			permissions[i] = fmt.Sprintf("%s:%s", p.ResourceCode, p.Code)
		}
	}
	return &models.UserDetails{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		ExternalUserID:  user.ExternalUserID,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		IsActive:        user.IsActive,
		IsEmailVerified: user.EmailVerifiedAt != nil,
		RoleName:        user.RoleName,
		RoleID:          user.RoleID,
		Permissions:     permissions,
	}, nil
}
