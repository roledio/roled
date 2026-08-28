package resource

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *resourceService) DeleteResource(ctx context.Context, req *models.DeleteResourceRequest) error {
	// Validate project access
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	if project.IsSystem {
		// System project resources cannot be deleted
		log.WithContext(ctx).Errorw("Deleting system project resources are not allowed")
		return pkgerrors.ErrOperationNotAvailable
	}

	// Find the resource
	resourceRepo := s.registry.ResourceRepository()
	resource, err := resourceRepo.FindByIDAndProjectID(ctx, req.ResourceID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find resource by ID and project ID", "error", err, "resource_id", req.ResourceID, "project_id", req.ProjectID)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if resource == nil {
		log.WithContext(ctx).Errorw("Resource not found", "resource_id", req.ResourceID, "project_id", req.ProjectID)
		return errors.ErrResourceNotFound
	}

	// Check if resource is default and prevent deletion
	if resource.IsDefault {
		log.WithContext(ctx).Errorw("Cannot delete default resource", "resource_id", req.ResourceID)
		return errors.ErrModifyDefaultResource
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Delete permissions associated with this resource
		permissionRepo := registry.PermissionRepository()
		_, err := permissionRepo.DeleteByResourceID(ctx, req.ResourceID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete permissions by resource ID", "error", err, "resource_id", req.ResourceID)
			return err
		}

		// Delete the resource
		resourceRepo := registry.ResourceRepository()
		affected, err := resourceRepo.Delete(ctx, resource)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete resource", "error", err, "resource_id", req.ResourceID)
			return err
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No resource deleted", "resource_id", req.ResourceID)
			return errors.ErrResourceNotFound
		}

		return nil
	})

	if err == nil {
		// Invalidate cache after successful deletion
		shared.InvalidateResourceCache(ctx, s.redisService, resource)
	}

	return err
}
