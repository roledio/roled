package role

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gookit/goutil/maputil"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/idutil"
)

func (s *roleService) CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.RoleDetailsAndPermissions, error) {
	account, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// No custom roles can be created for system project
		log.WithContext(ctx).Errorw("Creating a new role for system project is not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	code := strings.ToLower(req.Code)
	existingRole, err := s.registry.RoleRepository().FindByProjectIDAndCode(ctx, req.ProjectID, code)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find role by code and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existingRole != nil {
		log.WithContext(ctx).Errorw("Role already exists with the same code in the project", "project_id", req.ProjectID, "code", code)
		return nil, errors.ErrRoleCodeAlreadyUsed
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

	var newRole *entities.Role

	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Create role
		newRole = &entities.Role{
			ID:          idutil.NewID(),
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Code:        code,
			Name:        req.Name,
			Description: req.Description,
		}
		err = registry.RoleRepository().Create(ctx, newRole)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create role", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create role permissions
		if permissionLength > 0 {

			rolePermissions := make([]entities.RolePermission, 0, permissionLength)
			for _, permissionID := range req.PermissionIDs {
				rolePermissions = append(rolePermissions, entities.RolePermission{
					RoleID:       newRole.ID,
					PermissionID: permissionID,
				})
			}

			err = registry.RolePermissionRepository().Create(ctx, rolePermissions)
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

	now := time.Now().UTC()
	res := models.RoleDetailsAndPermissions{
		RoleDetails: models.RoleDetails{
			ID:          newRole.ID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Code:        newRole.Code,
			Name:        newRole.Name,
			Description: newRole.Description,
		},
		Permissions: resPermissions,
	}

	return &res, nil
}
