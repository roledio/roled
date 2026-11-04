package oauthconnection

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *oAuthConnectionService) UpdateOAuthConnection(ctx context.Context, req *models.UpdateOAuthConnectionRequest) (*models.OAuthConnectionDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem && req.CredentialType == constants.OAuthCredentialTypeDefault {
		log.WithContext(ctx).Warnw("Cannot update OAuth connection to default credentials for system project", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrDefaultCredentialsNotSupported
	}

	// Find existing connection
	oauthConnectionRepo := s.registry.OAuthConnectionRepository()
	existing, err := oauthConnectionRepo.FindByProjectIDAndProvider(ctx, req.ProjectID, req.Provider)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find OAuth connection", "error", err, "project_id", req.ProjectID, "provider", req.Provider)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existing == nil {
		log.WithContext(ctx).Warnw("OAuth connection not found", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrOAuthConnectionNotFound
	}

	// Make sure the system project already has an OAuth connection configured before
	// switching a non-system project's credential type to default.
	// (Whether the system connection is enabled can change at any time; we only
	// require it to be configured so the provider is known.)
	if !project.IsSystem && req.CredentialType == constants.OAuthCredentialTypeDefault {
		systemProject, err := s.registry.ProjectRepository().FindSystem(ctx)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find system project", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if systemProject == nil {
			log.WithContext(ctx).Warnw("System project does not exist", "project_id", req.ProjectID, "provider", req.Provider)
			return nil, pkgerrors.ErrSystemError.WithDebugMessage("System project does not exist")
		}
		systemOAuthConnection, err := oauthConnectionRepo.FindByProjectIDAndProvider(ctx, systemProject.ID, req.Provider)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to check system OAuth connection", "error", err, "project_id", req.ProjectID, "provider", req.Provider)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if systemOAuthConnection == nil {
			log.WithContext(ctx).Warnw("System project does not have OAuth connection configured for this provider", "project_id", req.ProjectID, "provider", req.Provider)
			return nil, errors.ErrDefaultCredentialsNotConfigured
		}
	}

	existing.Enabled = req.Enabled != nil && *req.Enabled
	existing.CredentialType = req.CredentialType

	if req.CredentialType == constants.OAuthCredentialTypeDefault {
		existing.ClientID = nil
		existing.ClientSecretEncrypted = nil
		existing.Scopes = nil
	} else {

		// Convert scopes to space-separated string
		scopesStr := strings.Join(req.Scopes, " ")

		// Encrypt client secret
		secretEncrypted, err := s.encryptOAuthClientSecret(ctx, req.ClientSecret)
		if err != nil {
			return nil, err
		}

		// Set client ID, encrypted secret, and scopes
		existing.ClientSecretEncrypted = &secretEncrypted
		existing.ClientID = &req.ClientID
		existing.Scopes = &scopesStr
	}

	affected, err := s.registry.OAuthConnectionRepository().Update(ctx, existing)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update OAuth connection", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No rows affected when updating OAuth connection", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrOAuthConnectionNotFound
	}

	// Invalidate cache
	shared.InvalidateOAuthConnectionCache(ctx, s.redisService, req.ProjectID, req.Provider)

	res := &models.OAuthConnectionDetails{
		ID:             existing.ID,
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      time.Now().UTC(),
		ProjectID:      existing.ProjectID,
		Provider:       existing.Provider,
		CredentialType: existing.CredentialType,
		Enabled:        existing.Enabled,
	}

	return res, nil
}
