package web

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/views"
	"github.com/roledio/roled/pkg/utils/flashutil"
	"github.com/roledio/roled/pkg/utils/requestutil"
)

func (h *handler) renderForgotPassword(c fiber.Ctx) error {
	ctx := c.Context()
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
	}
	var flash models.SubmitForgotPasswordFlash
	err := flashutil.ReadData(c, &flash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to read flash", "error", err)
	}
	templateData.Flash = flash
	req := models.RenderForgotPasswordRequest{}
	err = requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		return views.RenderTemplate(c, "templates/forgot-password", templateData.WithError(err), h.defaultConfig)
	}
	res, err := h.userService.RenderForgotPassword(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to render forgot password", "error", err)
		return views.RenderTemplate(c, "templates/forgot-password", templateData.WithError(err), h.defaultConfig)
	}
	templateData.Data = res
	templateData.Map = map[string]any{
		"CSRFToken":   csrf.TokenFromContext(c),
		"ClientID":    req.ClientID,
		"RedirectURI": req.RedirectURI,
		"LoginURL":    res.LoginURL,
	}
	return views.RenderTemplate(c, "templates/forgot-password", &templateData, h.defaultConfig)
}

func (h *handler) submitForgotPassword(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.SubmitForgotPasswordRequest{}
	err := requestutil.BindAndValidate(c, &req)
	redirectPath := fmt.Sprintf("/password/forgot?client_id=%s&redirect_uri=%s", req.ClientID, req.RedirectURI)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		flash := new(models.SubmitForgotPasswordFlash)
		flash.WithError(err)
		flash.Email = req.Email
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	err = h.userService.SubmitForgotPassword(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to submit forgot password", "error", err)
		flash := new(models.SubmitForgotPasswordFlash)
		flash.WithError(err)
		flash.Email = req.Email
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	flash := new(models.SubmitForgotPasswordFlash)
	flash.IsSuccess = true
	flashutil.SetData(c, flash)
	return c.Redirect().To(redirectPath)
}

func (h *handler) renderResetPassword(c fiber.Ctx) error {
	ctx := c.Context()
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
	}
	var flash models.SubmitResetPasswordFlash
	err := flashutil.ReadData(c, &flash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to read flash", "error", err)
	}
	templateData.Flash = flash
	req := models.RenderResetPasswordRequest{}
	err = requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		return views.RenderTemplate(c, "templates/reset-password", templateData.WithError(err), h.defaultConfig)
	}
	// Populate ProjectID from flash if available
	if flash.ProjectID != "" {
		req.ProjectID = &flash.ProjectID
	}
	result, err := h.userService.RenderResetPassword(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to render reset password page", "error", err)
		return views.RenderTemplate(c, "templates/reset-password", templateData.WithError(err), h.defaultConfig)
	}
	templateData.Data = result
	templateData.Map = map[string]any{
		"CSRFToken": csrf.TokenFromContext(c),
	}
	return views.RenderTemplate(c, "templates/reset-password", &templateData, h.defaultConfig)
}

func (h *handler) submitResetPassword(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.SubmitResetPasswordRequest{}
	err := requestutil.BindAndValidate(c, &req)
	redirectPath := fmt.Sprintf("/password/reset/%s", req.Token)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		flash := new(models.SubmitResetPasswordFlash)
		flash.WithError(err)
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	res, err := h.userService.SubmitResetPassword(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to submit reset password", "error", err)
		flash := new(models.SubmitResetPasswordFlash)
		flash.WithError(err)
		flashutil.SetData(c, flash)
		return c.Redirect().To(redirectPath)
	}
	templateData := models.TemplateData{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
		Data:      res,
	}
	return views.RenderTemplate(c, "templates/reset-password-success", &templateData, h.defaultConfig)
}
