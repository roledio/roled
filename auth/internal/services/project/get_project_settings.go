package project

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/constants/singeflightkeys"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/singleflightutil"
)

func (s *projectService) GetProjectSettings(ctx context.Context, req *models.GetProjectSettingsRequest) (*models.ProjectSettings, error) {
	key := singeflightkeys.GetProjectSettings(req.ProjectID)
	res, err, _ := singleflightutil.Do(&s.sfGroup, key, func() (*models.ProjectSettings, error) {
		return s.getProjectSettings(ctx, req)
	})
	return res, err
}

func (s *projectService) getProjectSettings(ctx context.Context, req *models.GetProjectSettingsRequest) (*models.ProjectSettings, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	setting, err := s.registry.ProjectSettingRepository().FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project settings by project ID", "error", err, "project_id", project.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if setting == nil {
		// Should not happen since project settings are created with default values when a project is created.
		log.WithContext(ctx).Errorw("Project setting not found", "project_id", project.ID)
		return nil, errors.ErrProjectSettingsNotFound
	}

	return &models.ProjectSettings{
		IsSignupEnabled:         setting.IsSignupEnabled,
		DefaultSignupRoleID:     setting.DefaultSignupRoleID,
		IsSignupVerifyEmail:     setting.IsSignupVerifyEmail,
		IsForgotPasswordEnabled: setting.IsForgotPasswordEnabled,
		IsAllowTempEmail:        setting.IsAllowTempEmail,
	}, nil
}
