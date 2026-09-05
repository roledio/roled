package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/flashutil"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
)

func (h *handler) handleGoogleOAuth(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.GoogleOAuthRequest{}
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		flash := models.SubmitAuthorizeFlash{}
		flash.WithError(err)
		flash.IsSignup = req.IsSignup
		flashutil.SetData(c, flash)
		return c.Redirect().To(h.buildRedirectURL(&req.RenderAuthorizeRequest, ""))
	}
	googleAuthURL, err := h.authorizeService.InitiateGoogleOAuth(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to initiate Google OAuth", "error", err)
		flash := new(models.SubmitAuthorizeFlash)
		flash.WithError(err)
		flash.IsSignup = req.IsSignup
		flashutil.SetData(c, flash)
		return c.Redirect().To(h.buildRedirectURL(&req.RenderAuthorizeRequest, ""))
	}
	return c.Redirect().To(googleAuthURL)
}

// handleGoogleOAuthCallback handles the Google OAuth callback.
func (h *handler) handleGoogleOAuthCallback(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.GoogleOAuthCallbackRequest{}
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate Google OAuth callback request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	redirectURL, err := h.authorizeService.HandleGoogleOAuthCallback(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to handle Google OAuth callback", "error", err)
		// Redirect to authorize page with error
		return c.Redirect().To("/authorize?error=google_oauth_failed")
	}

	return c.Redirect().To(redirectURL)
}
