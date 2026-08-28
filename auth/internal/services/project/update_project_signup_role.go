package project

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"go.openly.dev/pointy"
)

func (s *projectService) UpdateProjectSignupRole(ctx context.Context, req *models.UpdateProjectSignupRoleRequest) (*models.UpdateProjectSignupRoleResponse, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.IsSystem {
		log.WithContext(ctx).Errorw("Update signup role for system project is not allowed", "project_id", req.ProjectID)
		return nil, pkgerrors.ErrOperationNotAvailable.WithDebugMessage("Unable to update signup role for system project")
	}

	setting, err := s.registry.ProjectSettingRepository().FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project settings by project ID", "error", err, "project_id", project.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if setting == nil {
		log.WithContext(ctx).Errorw("Project setting not found", "project_id", project.ID)
		return nil, errors.ErrProjectSettingsNotFound
	}

	// Business rule: signup must be enabled before updating the default signup role.
	if !setting.IsSignupEnabled {
		log.WithContext(ctx).Errorw("Cannot update default signup role when signup is disabled", "project_id", project.ID)
		return nil, errors.ErrSignupMustBeEnabled
	}

	role, err := s.registry.RoleRepository().FindByIDAndProjectID(ctx, req.RoleID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find role by ID and project ID", "error", err, "role_id", req.RoleID, "project_id", project.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if role == nil {
		log.WithContext(ctx).Errorw("Default signup role not found by ID and project ID", "role_id", req.RoleID, "project_id", project.ID)
		return nil, errors.ErrRoleNotFound
	}

	if pointy.PointerValueEqual(setting.DefaultSignupRoleID, req.RoleID) {
		log.WithContext(ctx).Warnw("Default signup role is already set to the requested role", "project_id", project.ID, "role_id", req.RoleID)
		return &models.UpdateProjectSignupRoleResponse{
			RoleID:   role.ID,
			RoleName: role.Name,
		}, nil
	}

	setting.DefaultSignupRoleID = &req.RoleID

	affected, err := s.registry.ProjectSettingRepository().Update(ctx, setting)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update project setting", "error", err, "project_id", project.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No project setting updated", "project_id", project.ID)
		return nil, errors.ErrProjectSettingsNotFound
	}

	// Invalidate project setting cache after successful update
	shared.InvalidateProjectSettingCache(ctx, s.redis, project.ID)

	return &models.UpdateProjectSignupRoleResponse{
		RoleID:   role.ID,
		RoleName: role.Name,
	}, nil
}
