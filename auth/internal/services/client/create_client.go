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
	"github.com/roledio/roled/auth/internal/services/shared"
	"github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
)

func (s *clientService) CreateClient(ctx context.Context, req *models.CreateClientRequest) (*models.ClientDetailsAndPermissions, error) {
	account, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem {
		// No additional clients can be created for system project
		log.WithContext(ctx).Errorw("Creating a new client for system project is not allowed")
		return nil, pkgerrors.ErrOperationNotAvailable
	}

	resPermissions := []models.ClientPermission{}

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
		for _, p := range permissions {
			resPermissions = append(resPermissions, models.ClientPermission{
				ID:             p.ID,
				ResourceName:   p.ResourceName,
				PermissionName: p.Name,
			})
		}
	}

	var newClient *entities.Client

	err = s.registry.Tx(func(registry repositories.Registry) error {

		secretEncrypted, err := s.generateClientSecret(ctx)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to generate client secret", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create client
		newClient = &entities.Client{
			ID:              idutil.NewID(),
			AccountID:       account.ID,
			ProjectID:       req.ProjectID,
			Name:            req.Name,
			Description:     req.Description,
			SecretEncrypted: secretEncrypted,
			IsActive:        true,
			IsDefault:       false,
		}
		err = registry.ClientRepository().Create(ctx, newClient)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create client", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create client permissions
		if permissionLength > 0 {

			clientPermissions := make([]entities.ClientPermission, 0, permissionLength)
			for _, permissionID := range req.PermissionIDs {
				clientPermissions = append(clientPermissions, entities.ClientPermission{
					ClientID:     newClient.ID,
					PermissionID: permissionID,
				})
			}

			err = registry.ClientPermissionRepository().Create(ctx, clientPermissions)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to create client permissions", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	res := models.ClientDetailsAndPermissions{
		ClientDetails: models.ClientDetails{
			ID:          newClient.ID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Name:        newClient.Name,
			Description: newClient.Description,
			IsDefault:   newClient.IsDefault,
			IsActive:    newClient.IsActive,
		},
		Permissions: resPermissions,
	}

	return &res, nil
}

func (s *clientService) generateClientSecret(ctx context.Context) (string, error) {
	purpose := constants.KeyPurposeClientSecret
	derivedKey, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to derive key for client secret encryption", "error", err)
		return "", err
	}
	secret := idutil.NanoID(64)
	secretEncrypted, err := encryptionutil.EncryptAES(secret, derivedKey, purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to encrypt client secret", "error", err)
		return "", err
	}
	return secretEncrypted, nil
}
