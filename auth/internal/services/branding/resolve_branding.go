package branding

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
)

// ResolveBranding is for trusted web flows after their project has been validated.
// API callers must use GetBranding, which checks account ownership first.
func (s *service) ResolveBranding(ctx context.Context, project *entities.Project) (*models.BrandingDetails, error) {
	repo := s.registry.BrandingRepository()
	b, err := repo.FindByProjectID(ctx, project.ID)
	if err != nil {
		return nil, errors.ErrSystemError.WithError(err)
	}
	result := &models.BrandingDetails{
		ProjectID:       project.ID,
		SourceProjectID: project.ID,
		PrimaryColor:    "#ba8d1c",
		Rounding:        "small",
		EnableShadow:    true,
		LogoURL:         project.LogoURL,
	}
	if b == nil {
		result.IsDefault = true
		system, err := s.registry.ProjectRepository().FindSystem(ctx)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find system project", "error", err)
			return nil, errors.ErrSystemError.WithError(err)
		}
		if system == nil {
			log.WithContext(ctx).Error("System project not found!")
			return nil, errors.ErrSystemError.WithDebugMessage("System project not found")
		}
		result.SourceProjectID = system.ID
		b, err = repo.FindByProjectID(ctx, system.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find system project branding", "error", err)
			return nil, errors.ErrSystemError.WithError(err)
		}
	}
	if b != nil {
		if result.LogoURL == nil {
			result.LogoURL = b.LogoURL
		}
		result.FaviconURL = b.FaviconURL
		result.PrimaryColor = b.PrimaryColor
		result.Rounding = b.Rounding
		result.EnableShadow = b.EnableShadow
		result.EnableBorder = b.EnableBorder
	}
	return result, nil
}
