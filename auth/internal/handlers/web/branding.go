package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
)

func (h *handler) applyBranding(c fiber.Ctx, data *models.TemplateData, project *entities.Project) error {
	if project == nil {
		return errors.ErrSystemError.WithDebugMessage("Missing project for branding")
	}
	branding, err := h.brandingService.ResolveBranding(c.Context(), project.ID)
	if err != nil {
		return err
	}
	data.Branding = branding
	return nil
}
