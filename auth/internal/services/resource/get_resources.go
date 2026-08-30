package resource

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *resourceService) GetResources(ctx context.Context, req *models.GetResourcesRequest) ([]models.ResourceDetails, int, error) {
	results := []models.ResourceDetails{}
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return results, 0, err
	}
	resourceRepo := s.registry.ResourceRepository()
	count, err := resourceRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count resources", "error", err)
		return results, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return results, 0, nil
	}
	resources, err := resourceRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find resources", "error", err)
		return results, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	resourceIDs := make([]string, len(resources))
	for i, r := range resources {
		resourceIDs[i] = r.ID
	}
	permissionsRepo := s.registry.PermissionRepository()
	permissions, err := permissionsRepo.FindByResourceIDsAndSearch(ctx, resourceIDs, req.Search)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find permissions by resource IDs", "error", err)
		return results, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	mapPermissionsByResourceID := make(map[string][]models.Permission)
	for _, p := range permissions {
		mapPermissionsByResourceID[p.ResourceID] = append(mapPermissionsByResourceID[p.ResourceID], models.Permission{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
			IsDefault:   p.IsDefault,
		})
	}
	for _, r := range resources {
		res := models.ResourceDetails{
			ID:          r.ID,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
			IsDefault:   r.IsDefault,
			Permissions: mapPermissionsByResourceID[r.ID],
		}
		results = append(results, res)
	}
	return results, count, nil
}
