package resource

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/idutil"
)

func (s *resourceService) UpdateResource(ctx context.Context, req *models.UpdateResourceRequest) (*models.ResourceDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// System project resources cannot be updated
		log.WithContext(ctx).Errorw("Updating system project resources are not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	resource, err := s.registry.ResourceRepository().FindByIDAndProjectID(ctx, req.ResourceID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find resource by ID and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if resource == nil {
		log.WithContext(ctx).Errorw("Resource not found by ID and project ID",
			"resource_id", req.ResourceID,
			"project_id", req.ProjectID)
		return nil, errors.ErrResourceNotFound
	}
	if resource.IsDefault {
		log.WithContext(ctx).Errorw("Default resource cannot be updated", "resource_id", req.ResourceID)
		return nil, errors.ErrModifyDefaultResource
	}
	if resource.Code != req.Code {
		existingResource, err := s.registry.ResourceRepository().FindByProjectIDAndCode(ctx, req.ProjectID, req.Code)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find resource by project ID and code", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if existingResource != nil {
			log.WithContext(ctx).Errorw("Resource code already used in the project",
				"resource_code", req.Code,
				"project_id", req.ProjectID)
			return nil, errors.ErrResourceCodeAlreadyUsed
		}
	}

	var resPermissions []models.Permission
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Update resource
		resource.Name = req.Name
		resource.Code = req.Code
		resource.Description = req.Description
		affected, err := registry.ResourceRepository().Update(ctx, resource)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update resource", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No resource updated",
				"resource_id", req.ResourceID,
				"project_id", req.ProjectID)
			return errors.ErrResourceNotFound
		}

		// Delete existing permissions by resource ID
		permissionRepo := registry.PermissionRepository()
		_, err = permissionRepo.DeleteByResourceID(ctx, resource.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete permissions by resource ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		if len(req.Permissions) > 0 {
			// Create new permissions
			permissions := make([]entities.Permission, 0, len(req.Permissions))
			for _, p := range req.Permissions {
				permission := entities.Permission{
					ID:          idutil.NewID(),
					ResourceID:  resource.ID,
					Name:        p.Name,
					Code:        p.Code,
					Description: p.Description,
					IsDefault:   false,
				}
				permissions = append(permissions, permission)

				// Prepare response permissions
				resPermissions = append(resPermissions, models.Permission{
					ID:          permission.ID,
					Name:        permission.Name,
					Code:        permission.Code,
					Description: permission.Description,
					IsDefault:   permission.IsDefault,
				})
			}
			_, err = permissionRepo.Create(ctx, permissions)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to create permissions", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate cache after successful update
	shared.InvalidateResourceCache(ctx, s.redisService, resource)

	result := &models.ResourceDetails{
		ID:          resource.ID,
		CreatedAt:   resource.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
		Name:        resource.Name,
		Code:        resource.Code,
		Description: resource.Description,
		IsDefault:   resource.IsDefault,
		Permissions: resPermissions,
	}

	return result, nil
}
