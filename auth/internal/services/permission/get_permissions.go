package permission

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *permissionService) GetPermissions(ctx context.Context, req *models.GetPermissionsRequest) ([]models.PermissionDetails, int, error) {
	results := []models.PermissionDetails{}
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return results, 0, err
	}
	permissionRepo := s.registry.PermissionRepository()
	count, err := permissionRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count permissions", "error", err)
		return results, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return results, 0, nil
	}
	permissions, err := permissionRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find permissions", "error", err)
		return results, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	for _, p := range permissions {
		res := models.PermissionDetails{
			ID:           p.ID,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
			ResourceID:   p.ResourceID,
			ResourceName: p.ResourceName,
			ResourceCode: p.ResourceCode,
			Name:         p.Name,
			Code:         p.Code,
			Description:  p.Description,
			IsDefault:    p.IsDefault,
		}
		results = append(results, res)
	}
	return results, count, nil
}
