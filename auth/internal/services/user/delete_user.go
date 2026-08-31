package user

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *userService) DeleteUser(ctx context.Context, req *models.DeleteUserRequest) error {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	if project.IsSystem {
		// System project users cannot be deleted
		log.WithContext(ctx).Errorw("Deleting system project users are not allowed")
		return pkgerrors.ErrOperationNotAvailable
	}

	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByIDAndProjectID(ctx, req.UserID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID and project ID", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found by ID and project ID",
			"user_id", req.UserID,
			"project_id", project.ID)
		return errors.ErrUserNotFound
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {
		// 1. Delete access tokens associated with the user
		_, err = registry.AccessTokenRepository().DeleteByUserID(ctx, user.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete access tokens by user ID", "error", err, "user_id", user.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// 2. Delete user role associations
		_, err = registry.UserRoleRepository().DeleteByUserID(ctx, user.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete user roles by user ID", "error", err, "user_id", user.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// 3. Delete user record
		affected, err := registry.UserRepository().DeleteByID(ctx, user.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete user", "error", err, "user_id", user.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No rows affected when deleting user", "user_id", user.ID)
			return errors.ErrUserNotFound
		}

		return nil
	})

	if err == nil {
		// Invalidate cache after successful deletion
		shared.InvalidateUserCache(ctx, s.redisService, user, &shared.OldUserCacheKeyParts{
			Email:          user.Email,
			ExternalUserID: user.ExternalUserID,
		})
	}

	return err
}
