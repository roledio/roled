package oauthconnection

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *oAuthConnectionService) DeleteOAuthConnection(ctx context.Context, req *models.DeleteOAuthConnectionRequest) error {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	affected, err := s.registry.OAuthConnectionRepository().Delete(ctx, req.ProjectID, req.Provider)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to delete OAuth connection", "error", err, "project_id", req.ProjectID, "provider", req.Provider)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No OAuth connection deleted by project ID and provider", "project_id", req.ProjectID, "provider", req.Provider)
		return errors.ErrOAuthConnectionNotFound
	}

	// Invalidate cache
	shared.InvalidateOAuthConnectionCache(ctx, s.redisService, req.ProjectID, req.Provider)

	return nil
}
