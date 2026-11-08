package oauthconnection

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
)

func (s *oAuthConnectionService) CreateOAuthConnection(ctx context.Context, req *models.CreateOAuthConnectionRequest) (*models.OAuthConnectionDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.IsSystem && req.CredentialType == constants.OAuthCredentialTypeDefault {
		log.WithContext(ctx).Warnw("Cannot create default OAuth connection for system project", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrDefaultCredentialsNotSupported
	}

	// Check if connection already exists
	oauthConnectionRepo := s.registry.OAuthConnectionRepository()
	existing, err := oauthConnectionRepo.FindByProjectIDAndProvider(ctx, req.ProjectID, req.Provider)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to check existing OAuth connection", "error", err, "project_id", req.ProjectID, "provider", req.Provider)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existing != nil {
		log.WithContext(ctx).Warnw("OAuth connection already exists", "project_id", req.ProjectID, "provider", req.Provider)
		return nil, errors.ErrOAuthConnectionAlreadyExists
	}

	newConnection := &entities.OAuthConnection{
		ID:             idutil.NewID(),
		ProjectID:      req.ProjectID,
		Provider:       req.Provider,
		CredentialType: req.CredentialType,
		Enabled:        req.Enabled != nil && *req.Enabled,
	}

	if !project.IsSystem && req.CredentialType == constants.OAuthCredentialTypeDefault {
		if err := s.validateSystemOAuthConnectionConfigured(ctx, req.ProjectID, req.Provider); err != nil {
			return nil, err
		}
	}

	if req.CredentialType == constants.OAuthCredentialTypeCustom {
		if err := s.setCustomOAuthCredentials(ctx, newConnection, req.ClientID, req.ClientSecret, req.Scopes); err != nil {
			return nil, err
		}
	}

	err = oauthConnectionRepo.Create(ctx, newConnection)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to create OAuth connection", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Invalidate the oauth connection list for this project
	shared.InvalidateOAuthConnectionCache(ctx, s.redisService, req.ProjectID, req.Provider)

	now := time.Now().UTC()
	res := &models.OAuthConnectionDetails{
		ID:             newConnection.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
		ProjectID:      newConnection.ProjectID,
		Provider:       newConnection.Provider,
		CredentialType: newConnection.CredentialType,
		Enabled:        newConnection.Enabled,
	}

	return res, nil
}
