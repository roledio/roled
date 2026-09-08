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
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
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
