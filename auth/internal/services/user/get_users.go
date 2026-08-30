package user

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *userService) GetUsers(ctx context.Context, req *models.GetUsersRequest) ([]models.UserDetails, int, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, 0, err
	}

	userRepo := s.registry.UserRepository()
	count, err := userRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count users", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}

	users, err := userRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find users", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	var resp []models.UserDetails
	for _, user := range users {
		resp = append(resp, models.UserDetails{
			ID:              user.ID,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
			Email:           user.Email,
			ExternalUserID:  user.ExternalUserID,
			DisplayName:     user.DisplayName,
			AvatarURL:       user.AvatarURL,
			IsActive:        user.IsActive,
			IsEmailVerified: user.EmailVerifiedAt != nil,
			RoleID:          user.RoleID,
			RoleName:        user.RoleName,
		})
	}

	return resp, count, nil
}
