package client

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *clientService) DeleteClient(ctx context.Context, req *models.DeleteClientRequest) error {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	if project.IsSystem {
		// System project clients cannot be deleted
		log.WithContext(ctx).Errorw("Deleting system project clients are not allowed")
		return pkgerrors.ErrOperationNotAvailable
	}

	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByIDAndProjectID(ctx, req.ClientID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by ID and project ID", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found by ID and project ID",
			"client_id", req.ClientID,
			"project_id", project.ID)
		return errors.ErrClientNotFound
	}
	if client.IsDefault {
		log.WithContext(ctx).Errorw("Default client cannot be deleted", "client_id", client.ID)
		return errors.ErrDeleteDefaultClient
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Delete access tokens associated with the client
		_, err := registry.AccessTokenRepository().DeleteByClientID(ctx, client.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete access tokens by client ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Delete client permissions
		_, err = registry.ClientPermissionRepository().DeleteByClientID(ctx, client.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete permissions by client ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Delete the client
		affected, err := registry.ClientRepository().Delete(ctx, client)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete client by ID", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No client deleted by ID", "client_id", client.ID)
			return errors.ErrClientNotFound
		}
		return nil
	})

	if err == nil {
		// Invalidate cache after successful deletion
		shared.InvalidateClientCache(ctx, s.redisService, client)
		shared.InvalidateClientPermissionsCache(ctx, s.redisService, client.ID)
	}

	return err
}
