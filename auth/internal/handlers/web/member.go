package web

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/views"
	"github.com/roledio/roled/auth/pkg/utils/flashutil"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
)

func (h *handler) renderActivateMember(c fiber.Ctx) error {
	ctx := c.Context()
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
	}
	var flash models.SubmitActivateMemberFlash
	err := flashutil.ReadData(c, &flash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to read flash", "error", err)
	}
	templateData.Flash = flash
	req := models.RenderActivateMemberRequest{}
	err = requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		return views.RenderTemplate(c, "templates/activate-member", templateData.WithError(err), h.defaultConfig)
	}
	// Populate UserID from flash if available
	if flash.UserID != "" {
		req.UserID = &flash.UserID
	}
	res, err := h.memberService.RenderActivateMember(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to render activate member", "error", err)
		return views.RenderTemplate(c, "templates/activate-member", templateData.WithError(err), h.defaultConfig)
	}
	templateData.Data = res
	templateData.Map = map[string]any{
		"CSRFToken": csrf.TokenFromContext(c),
	}
	return views.RenderTemplate(c, "templates/activate-member", &templateData, h.defaultConfig)
}

func (h *handler) submitActivateMember(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.SubmitActivateMemberRequest{}
	err := requestutil.BindAndValidate(c, &req)
	redirectPath := fmt.Sprintf("/member/activate/%s", req.Token)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		flash := new(models.SubmitActivateMemberFlash)
		flash.WithError(err)
		flash.DisplayName = req.DisplayName
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	res, err := h.memberService.SubmitActivateMember(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to submit activate member", "error", err)
		flash := new(models.SubmitActivateMemberFlash)
		flash.WithError(err)
		flash.DisplayName = req.DisplayName
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
		Data:      res,
	}
	return views.RenderTemplate(c, "templates/activate-member-success", &templateData, h.defaultConfig)
}
