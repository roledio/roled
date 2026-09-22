package branding

import (
	"context"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
)

func (s *service) GetBranding(ctx context.Context, req *models.GetBrandingRequest) (*models.BrandingDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	return s.ResolveBranding(ctx, project.ID)
}
