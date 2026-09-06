package oauthconnection

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/copyutil"
)

func (s *oAuthConnectionService) GetOAuthConnections(ctx context.Context, req *models.GetOAuthConnectionsRequest) ([]models.OAuthConnectionDetails, error) {
	_, _, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	oauthConnectionRepo := s.registry.OAuthConnectionRepository()
	connections, err := oauthConnectionRepo.FindByProjectID(ctx, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project OAuth connections", "error", err, "project_id", req.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	results := []models.OAuthConnectionDetails{}
	err = copyutil.Copy(connections, &results)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to copy OAuth connections", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	return results, nil
}
