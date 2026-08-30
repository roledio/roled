package role

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gookit/goutil/maputil"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *roleService) UpdateRole(ctx context.Context, req *models.UpdateRoleRequest) (*models.RoleDetailsAndPermissions, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// System project roles cannot be updated
		log.WithContext(ctx).Errorw("Updating system project roles are not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	roleRepo := s.registry.RoleRepository()
	role, err := roleRepo.FindByIDAndProjectID(ctx, req.RoleID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find role by ID and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if role == nil {
		log.WithContext(ctx).Errorw("Role not found for given ID and project ID",
			"role_id", req.RoleID,
			"project_id", req.ProjectID)
		return nil, errors.ErrRoleNotFound
	}

	code := strings.ToLower(req.Code)
	if code != role.Code {
		// Role code is being updated, check if other roles in the same project have the same code
		existingRole, err := roleRepo.FindByProjectIDAndCode(ctx, req.ProjectID, code)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find role by code and project ID", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if existingRole != nil {
			log.WithContext(ctx).Errorw("Role with the same code already exists in the project",
				"code", code,
				"project_id", req.ProjectID)
			return nil, errors.ErrRoleCodeAlreadyUsed
		}
	}

	resPermissions := []models.RolePermission{}

	permissionLength := len(req.PermissionIDs)

	if permissionLength > 0 {

		permissionRepo := s.registry.PermissionRepository()
		permissions, err := permissionRepo.FindByIDs(ctx, req.PermissionIDs)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find permissions by IDs", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if len(permissions) != len(req.PermissionIDs) {
			mapNotFoundIDs := make(map[string]bool)
			for _, id := range req.PermissionIDs {
				mapNotFoundIDs[id] = true
			}
			for _, p := range permissions {
				delete(mapNotFoundIDs, p.ID)
			}
			notFoundIDs := maputil.Keys(mapNotFoundIDs)
			log.WithContext(ctx).Errorw("Some permissions not found for given IDs", "not_found_ids", notFoundIDs)
			return nil, errors.ErrSomePermissionsNotFound(notFoundIDs)
		}
		for _, p := range permissions {
			resPermissions = append(resPermissions, models.RolePermission{
				ID:             p.ID,
				ResourceName:   p.ResourceName,
				PermissionName: p.Name,
			})
		}
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Update role
		role.Code = code
		role.Name = req.Name
		role.Description = req.Description
		affected, err := registry.RoleRepository().Update(ctx, role)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update role", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No rows affected when updating role", "role_id", role.ID)
			return errors.ErrRoleNotFound
		}

		// Update role permissions
		rolePermissionRepo := registry.RolePermissionRepository()
		_, err = rolePermissionRepo.DeleteByRoleID(ctx, role.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete role permissions by role ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if permissionLength > 0 {
			rolePermissions := make([]entities.RolePermission, 0, permissionLength)
			for _, permissionID := range req.PermissionIDs {
				rolePermissions = append(rolePermissions, entities.RolePermission{
					RoleID:       role.ID,
					PermissionID: permissionID,
				})
			}
			err = rolePermissionRepo.Create(ctx, rolePermissions)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to create role permissions", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate cache after successful update
	shared.InvalidateRoleCache(ctx, s.redisService, s.registry, role)
	shared.InvalidateRolePermissionsCache(ctx, s.redisService, role.ID)

	res := &models.RoleDetailsAndPermissions{
		RoleDetails: models.RoleDetails{
			ID:          role.ID,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   time.Now().UTC(),
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
		},
		Permissions: resPermissions,
	}

	return res, nil
}
