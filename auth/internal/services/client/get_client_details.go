package client

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	"github.com/roledio/roled/pkg/constants"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/encryptionutil"
)

func (s *clientService) GetClientDetails(ctx context.Context, req *models.GetClientDetailsRequest) (*models.ClientDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByIDAndProjectID(ctx, req.ClientID, req.ProjectID)
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

	return s.buildClientDetailsWithSecret(ctx, client)
}

func (s *clientService) buildClientDetailsWithSecret(ctx context.Context, client *entities.Client) (*models.ClientDetails, error) {
	purpose := constants.KeyPurposeClientSecret
	derivedKey, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to derive key for client secret encryption", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	secret, err := encryptionutil.DecryptAES(client.SecretEncrypted, derivedKey, purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to decrypt client secret", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	clientPermissions, err := s.registry.ClientPermissionRepository().FindByClientID(ctx, client.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find permissions for client", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	permissionIDs := make([]string, len(clientPermissions))
	for i, cp := range clientPermissions {
		permissionIDs[i] = cp.PermissionID
	}

	res := &models.ClientDetails{
		ID:            client.ID,
		CreatedAt:     client.CreatedAt,
		UpdatedAt:     client.UpdatedAt,
		Name:          client.Name,
		Description:   client.Description,
		IsActive:      client.IsActive,
		IsDefault:     client.IsDefault,
		Secret:        secret,
		PermissionIDs: permissionIDs,
	}
	return res, nil
}
