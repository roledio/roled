package role

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"go.openly.dev/pointy"
)

func (s *roleService) GetRoles(ctx context.Context, req *models.GetProjectRolesRequest) ([]models.RoleDetails, int, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, 0, err
	}

	roleRepo := s.registry.RoleRepository()
	count, err := roleRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count roles", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}

	roles, err := roleRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find roles", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}

	setting, err := s.registry.ProjectSettingRepository().FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project settings", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if setting == nil {
		log.WithContext(ctx).Errorw("Project settings not found", "project_id", req.ProjectID)
		return nil, 0, errors.ErrProjectSettingsNotFound
	}

	var resp []models.RoleDetails
	for _, role := range roles {
		// Check if the role is the default signup role
		isDefaultSignup := pointy.PointerValueEqual(setting.DefaultSignupRoleID, role.ID)
		resp = append(resp, models.RoleDetails{
			ID:              role.ID,
			CreatedAt:       role.CreatedAt,
			UpdatedAt:       role.UpdatedAt,
			Code:            role.Code,
			Name:            role.Name,
			Description:     role.Description,
			IsDefaultSignup: &isDefaultSignup,
		})
	}

	return resp, count, nil
}
