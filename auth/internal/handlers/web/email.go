package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/views"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
)

func (h *handler) renderVerifyEmail(c fiber.Ctx) error {
	ctx := c.Context()
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
	}
	req := models.VerifyEmailRequest{}
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		return views.RenderTemplate(c, "templates/verify-email", templateData.WithError(err), h.defaultConfig)
	}
	res, err := h.userService.RenderVerifyEmail(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to render verify email", "error", err)
		return views.RenderTemplate(c, "templates/verify-email", templateData.WithError(err), h.defaultConfig)
	}
	templateData.Data = res
	return views.RenderTemplate(c, "templates/verify-email", &templateData, h.defaultConfig)
}
