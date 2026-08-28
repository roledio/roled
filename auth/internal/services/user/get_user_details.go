package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *userService) GetUserDetails(ctx context.Context, req *models.GetUserDetailsRequest) (*models.UserDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return s.processGetUserDetails(ctx, req)
}

func (s *userService) processGetUserDetails(ctx context.Context, req *models.GetUserDetailsRequest) (*models.UserDetails, error) {
	user, err := s.registry.UserRepository().FindByIDAndProjectIDJoinRole(ctx, req.UserID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID and project ID", "error", err)
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
