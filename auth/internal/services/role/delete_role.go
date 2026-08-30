package role

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *roleService) DeleteRole(ctx context.Context, req *models.DeleteRoleRequest) error {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	if project.IsSystem {
		// System project roles cannot be deleted
		log.WithContext(ctx).Errorw("Deleting system project roles are not allowed")
		return pkgerrors.ErrOperationNotAvailable
	}

	roleRepo := s.registry.RoleRepository()
	role, err := roleRepo.FindByIDAndProjectID(ctx, req.RoleID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find role by ID and project ID", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if role == nil {
		log.WithContext(ctx).Errorw("Role not found by ID and project ID",
			"role_id", req.RoleID,
			"project_id", project.ID)
		return errors.ErrRoleNotFound
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Delete role permissions
		_, err = registry.RolePermissionRepository().DeleteByRoleID(ctx, role.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete permissions by role ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Delete the role
		affected, err := registry.RoleRepository().DeleteByID(ctx, role.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete role by ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No role deleted by ID", "role_id", role.ID)
			return errors.ErrRoleNotFound
		}
		return nil
	})

	if err == nil {
		// Invalidate cache after successful deletion
		shared.InvalidateRoleCache(ctx, s.redisService, s.registry, role)
		shared.InvalidateRolePermissionsCache(ctx, s.redisService, role.ID)
	}

	return err
}
