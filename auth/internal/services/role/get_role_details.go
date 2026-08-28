package role

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *roleService) GetRoleDetails(ctx context.Context, req *models.GetRoleDetailsRequest) (*models.RoleDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
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

	rolePermissions, err := s.registry.ClientPermissionRepository().FindByRoleID(ctx, role.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find permissions for role", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	permissionIDs := make([]string, len(rolePermissions))
	for i, cp := range rolePermissions {
		permissionIDs[i] = cp.PermissionID
	}

	res := &models.RoleDetails{
		ID:            role.ID,
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
		Code:          role.Code,
		Name:          role.Name,
		Description:   role.Description,
		PermissionIDs: permissionIDs,
	}
	return res, nil
}
