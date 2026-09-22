package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getProjectBranding(c fiber.Ctx) error {
	var req models.GetBrandingRequest
	if err := requestutil.BindAndValidate(c, &req); err != nil {
		return responseutil.SendError(c, err)
	}
	branding, err := h.brandingService.GetBranding(c.Context(), &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, branding)
}

func (h *handler) updateProjectBranding(c fiber.Ctx) error {
	var req models.UpdateBrandingRequest
	if err := requestutil.BindAndValidate(c, &req); err != nil {
		return responseutil.SendError(c, err)
	}
	branding, err := h.brandingService.UpdateBranding(c.Context(), &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, branding)
}
