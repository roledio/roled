package client

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gookit/goutil/maputil"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"go.openly.dev/pointy"
)

func (s *clientService) UpdateClient(ctx context.Context, req *models.UpdateClientRequest) (*models.ClientDetailsAndPermissions, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// System project clients cannot be updated
		log.WithContext(ctx).Errorw("Updating system project clients are not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	client, err := s.registry.ClientRepository().FindByIDAndProjectID(ctx, req.ClientID, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by ID and project ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found for given ID and project ID",
			"client_id", req.ClientID,
			"project_id", req.ProjectID)
		return nil, errors.ErrClientNotFound
	}

	customPermissions := []interfaces.PermissionResource{}
	permissionLength := len(req.PermissionIDs)
	if permissionLength > 0 {

		permissionRepo := s.registry.PermissionRepository()
		permissions, err := permissionRepo.FindByIDs(ctx, req.PermissionIDs)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find permissions by IDs", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if len(permissions) != len(req.PermissionIDs) {
			mapNotFoundIDs := make(map[string]bool)
			for _, id := range req.PermissionIDs {
				mapNotFoundIDs[id] = true
			}
			for _, p := range permissions {
				delete(mapNotFoundIDs, p.ID)
			}
			notFoundIDs := maputil.Keys(mapNotFoundIDs)
			log.WithContext(ctx).Errorw("Some permissions not found for given IDs", "not_found_ids", notFoundIDs)
			return nil, errors.ErrSomePermissionsNotFound(notFoundIDs)
		}
		// Filter custom permissions (non-default) from request
		for _, p := range permissions {
			if !p.IsDefault {
				customPermissions = append(customPermissions, p)
			}
		}
	}

	isActive := pointy.BoolValue(req.IsActive, false)
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Update client
		client.Name = req.Name
		client.Description = req.Description
		client.IsActive = isActive
		affected, err := registry.ClientRepository().Update(ctx, client)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update client", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No rows affected when updating client", "client_id", client.ID)
			return errors.ErrClientNotFound
		}

		clientPermissionRepo := registry.ClientPermissionRepository()
		clientPermissions := []entities.ClientPermission{}
		if client.IsDefault {
			// Default client must have all default permissions of the project, so we need to get them
			// and append with custom permissions if provided in the request
			defaultPermissions, err := registry.PermissionRepository().FindByProjectID(ctx, client.ProjectID, new(true))
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to get default permissions for project", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
			// Build client permissions with default and custom permissions
			for _, p := range defaultPermissions {
				clientPermissions = append(clientPermissions, entities.ClientPermission{
					ClientID:     client.ID,
					PermissionID: p.ID,
				})
			}
			for _, p := range customPermissions {
				clientPermissions = append(clientPermissions, entities.ClientPermission{
					ClientID:     client.ID,
					PermissionID: p.ID,
				})
			}
		} else {
			// Non-default client can add any permissions
			for _, id := range req.PermissionIDs {
				clientPermissions = append(clientPermissions, entities.ClientPermission{
					ClientID:     client.ID,
					PermissionID: id,
				})
			}
		}

		// Delete existing client permissions
		_, err = clientPermissionRepo.DeleteByClientID(ctx, client.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete client permissions by client ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create new client permissions
		err = clientPermissionRepo.Create(ctx, clientPermissions)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create client permissions", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate cache after successful update
	shared.InvalidateClientCache(ctx, s.redisService, client)
	shared.InvalidateClientPermissionsCache(ctx, s.redisService, client.ID)

	// Build response permissions - fetch actual stored permissions from database
	resPermissions := []models.ClientPermission{}
	allPermissions, err := s.registry.PermissionRepository().FindByClientID(ctx, client.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to fetch client permissions for response", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	for _, p := range allPermissions {
		resPermissions = append(resPermissions, models.ClientPermission{
			ID:             p.ID,
			ResourceName:   p.ResourceName,
			PermissionName: p.Name,
		})
	}

	res := &models.ClientDetailsAndPermissions{
		ClientDetails: models.ClientDetails{
			ID:          client.ID,
			CreatedAt:   client.CreatedAt,
			UpdatedAt:   time.Now().UTC(),
			Name:        req.Name,
			Description: req.Description,
			IsDefault:   client.IsDefault,
			IsActive:    isActive,
		},
		Permissions: resPermissions,
	}

	return res, nil
}
