package resource

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
)

func (s *resourceService) CreateResource(ctx context.Context, req *models.CreateResourceRequest) (*models.ResourceDetails, error) {
	account, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// No custom resources can be created for system project
		log.WithContext(ctx).Errorw("Creating a new resource for system project is not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	existingResource, err := s.registry.ResourceRepository().FindByProjectIDAndCode(ctx, req.ProjectID, req.Code)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find resource by project ID and code", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existingResource != nil {
		log.WithContext(ctx).Errorw("Resource code already used in the project",
			"resource_code", req.Code,
			"project_id", req.ProjectID,
			"resource_id", existingResource.ID)
		return nil, errors.ErrResourceCodeAlreadyUsed
	}

	var resPermissions []models.Permission
	var resource *entities.Resource
	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Create resource
		resource = &entities.Resource{
			ID:          idutil.NewID(),
			AccountID:   account.ID,
			ProjectID:   project.ID,
			Name:        req.Name,
			Code:        req.Code,
			Description: req.Description,
			IsDefault:   false,
		}
		_, err = registry.ResourceRepository().Create(ctx, []entities.Resource{*resource})
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create resource", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		if len(req.Permissions) > 0 {
			// Create permissions
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
			_, err = registry.PermissionRepository().Create(ctx, permissions)
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

	now := time.Now().UTC()
	res := &models.ResourceDetails{
		ID:          resource.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        resource.Name,
		Code:        resource.Code,
		Description: resource.Description,
		IsDefault:   resource.IsDefault,
		Permissions: resPermissions,
	}
	return res, nil
}
