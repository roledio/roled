package resource

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *resourceService) GetResourceDetails(ctx context.Context, req *models.GetResourceDetailsRequest) (*models.ResourceDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	resourceRepo := s.registry.ResourceRepository()
	resource, err := resourceRepo.FindByIDAndProjectID(ctx, req.ResourceID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find resource by ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if resource == nil {
		log.WithContext(ctx).Errorw("Resource not found by ID", "resource_id", req.ResourceID)
		return nil, errors.ErrResourceNotFound
	}
	permissionsRepo := s.registry.PermissionRepository()
	permissions, err := permissionsRepo.FindByResourceIDsAndSearch(ctx, []string{req.ResourceID}, "")
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find permissions by resource ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	result := &models.ResourceDetails{
		ID:          resource.ID,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   resource.UpdatedAt,
		Name:        resource.Name,
		Code:        resource.Code,
		Description: resource.Description,
		IsDefault:   resource.IsDefault,
	}
	for _, p := range permissions {
		result.Permissions = append(result.Permissions, models.Permission{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
			IsDefault:   p.IsDefault,
		})
	}
	return result, nil
}
