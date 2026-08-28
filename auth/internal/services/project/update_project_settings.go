package project

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *projectService) UpdateProjectSettings(ctx context.Context, req *models.UpdateProjectSettingsRequest) (*models.ProjectSettings, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.IsSystem {
		log.WithContext(ctx).Errorw("Update project settings for system project is not allowed", "project_id", req.ProjectID)
		return nil, pkgerrors.ErrOperationNotAvailable.WithDebugMessage("Unable to update settings for system project")
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

	// Business rule: a default signup role is required when signup is enabled.
	if *req.IsSignupEnabled && req.DefaultSignupRoleID == nil {
		log.WithContext(ctx).Errorw("DefaultSignupRoleID is required when IsSignupEnabled is true", "project_id", project.ID)
		return nil, errors.ErrDefaultSignupRoleRequired
	}

	if req.DefaultSignupRoleID != nil {
		role, err := s.registry.RoleRepository().FindByIDAndProjectID(ctx, *req.DefaultSignupRoleID, project.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find role by ID and project ID", "error", err, "role_id", *req.DefaultSignupRoleID, "project_id", project.ID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if role == nil {
			log.WithContext(ctx).Errorw("Default signup role not found by ID and project ID", "role_id", *req.DefaultSignupRoleID, "project_id", project.ID)
			return nil, errors.ErrRoleNotFound
		}
	}

	setting.IsSignupEnabled = *req.IsSignupEnabled
	setting.IsForgotPasswordEnabled = *req.IsForgotPasswordEnabled
	if !setting.IsSignupEnabled {
		// Reset signup-related settings when signup is disabled
		setting.DefaultSignupRoleID = nil
		setting.IsSignupVerifyEmail = false
		setting.IsAllowTempEmail = false
	} else {
		// Set signup-related settings to requested values when signup is enabled
		setting.DefaultSignupRoleID = req.DefaultSignupRoleID
		setting.IsSignupVerifyEmail = *req.IsSignupVerifyEmail
		setting.IsAllowTempEmail = *req.IsAllowTempEmail
	}

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

	return &models.ProjectSettings{
		IsSignupEnabled:         setting.IsSignupEnabled,
		DefaultSignupRoleID:     setting.DefaultSignupRoleID,
		IsSignupVerifyEmail:     setting.IsSignupVerifyEmail,
		IsForgotPasswordEnabled: setting.IsForgotPasswordEnabled,
		IsAllowTempEmail:        setting.IsAllowTempEmail,
	}, nil
}
