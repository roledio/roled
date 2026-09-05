package web

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/views"
	"github.com/roledio/roled/auth/pkg/utils/flashutil"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
)

func (h *handler) renderAuthorize(c fiber.Ctx) error {
	ctx := c.Context()
	templateData := models.TemplateData{
		Config:    h.defaultConfig,
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
	}
	var flash models.SubmitAuthorizeFlash
	err := flashutil.ReadData(c, &flash)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to read flash", "error", err)
	}
	req := models.RenderAuthorizeRequest{}
	err = requestutil.BindAndValidate(c, &req)
	flash.IsSignup = flash.IsSignup || req.IsSignup
	templateData.Flash = flash
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		return views.RenderTemplate(c, "templates/authorize", templateData.WithError(err), h.defaultConfig)
	}
	result, err := h.authorizeService.RenderAuthorize(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to render authorize page", "error", err)
		return views.RenderTemplate(c, "templates/authorize", templateData.WithError(err), h.defaultConfig)
	}
	templateData.Data = result
	templateData.Map = map[string]any{
		"CSRFToken":          csrf.TokenFromContext(c),
		"Req":                req,
		"ForgotPasswordPath": h.buildForgotPasswordPath(&req),
		"GoogleOAuthPath":    h.buildGoogleOAuthPath(&req),
	}
	return views.RenderTemplate(c, "templates/authorize", &templateData, h.defaultConfig)
}

// submitAuthorize handles the submission of the authorization form.
// It processes the user's input and redirects accordingly.
// On error, it logs the error, sets flash, and redirects back to the authorize page.
// On success, it redirects to the provided redirect URI with the authorization code
// from AuthorizeService and the state from the original request.
func (h *handler) submitAuthorize(c fiber.Ctx) error {
	ctx := c.Context()
	req := models.SubmitAuthorizeRequest{}
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to bind and validate request", "error", err)
		flash := models.SubmitAuthorizeFlash{}
		flash.WithError(err)
		flash.Email = req.Email
		flash.IsSignup = req.IsSignup
		flashutil.SetData(c, flash)
		return c.Redirect().To(h.buildRedirectURL(&req.RenderAuthorizeRequest, ""))
	}
	// Process the authorization submission
	result, err := h.authorizeService.SubmitAuthorize(ctx, &req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to submit authorize", "error", err)
		flash := new(models.SubmitAuthorizeFlash)
		flash.WithError(err)
		flash.Email = req.Email
		flash.IsSignup = req.IsSignup
		flashutil.SetData(c, flash)
		return c.Redirect().To(h.buildRedirectURL(&req.RenderAuthorizeRequest, ""))
	}
	return c.Redirect().To(h.buildRedirectURL(&req.RenderAuthorizeRequest, result.Code))
}

func (h *handler) buildRedirectURL(req *models.RenderAuthorizeRequest, authCode string) string {
	if authCode == "" {
		query := url.Values{}
		query.Set("client_id", req.ClientID)
		query.Set("redirect_uri", req.RedirectURI)
		query.Set("response_type", req.ResponseType)
		query.Set("code_challenge", req.CodeChallenge)
		query.Set("code_challenge_method", req.CodeChallengeMethod)
		query.Set("state", req.State)
		query.Set("is_signup", fmt.Sprintf("%v", req.IsSignup))
		return fmt.Sprintf("/authorize?%s", query.Encode())
	}
	return fmt.Sprintf("%s?code=%s&state=%s",
		req.RedirectURI,
		authCode,
		req.State)
}

func (h *handler) buildForgotPasswordPath(req *models.RenderAuthorizeRequest) string {
	query := url.Values{}
	query.Set("client_id", req.ClientID)
	query.Set("redirect_uri", req.RedirectURI)
	return fmt.Sprintf("/password/forgot?%s", query.Encode())
}

func (h *handler) buildGoogleOAuthPath(req *models.RenderAuthorizeRequest) string {
	query := url.Values{}
	query.Set("client_id", req.ClientID)
	query.Set("redirect_uri", req.RedirectURI)
	query.Set("response_type", req.ResponseType)
	query.Set("code_challenge", req.CodeChallenge)
	query.Set("code_challenge_method", req.CodeChallengeMethod)
	query.Set("state", req.State)
	query.Set("is_signup", fmt.Sprintf("%v", req.IsSignup))
	return fmt.Sprintf("/oauth/google?%s", query.Encode())
}
