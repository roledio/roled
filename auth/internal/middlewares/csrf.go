package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v3/extractors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/flashutil"
)

func CSRF(defaultConfig *configs.DefaultConfig, redisService infra.RedisService) fiber.Handler {
	trustedOrigins := []string{}
	// Generate trusted origins based on CORS allowed domains, supporting both http and https schemes
	for _, domain := range defaultConfig.CORS.AllowedDomains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			trustedOrigins = append(trustedOrigins, "https://*."+domain)
			trustedOrigins = append(trustedOrigins, "http://*."+domain)
		}
	}
	return csrf.New(csrf.Config{
		Next: func(c fiber.Ctx) bool {
			isPathSystem := strings.HasPrefix(c.Path(), "/system")
			isPathAPI := strings.HasPrefix(c.Path(), "/api")
			return isPathAPI || isPathSystem // Skip CSRF for API routes and system routes
		},
		TrustedOrigins: trustedOrigins,
		Extractor:      extractors.FromForm("csrf_token"),
		Storage:        redisService,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			log.WithContext(c.Context()).Debugw("CSRF middleware error", "error", err)
			flash := new(models.SubmitAuthorizeFlash).WithError(errors.ErrInvalidCSRFToken.WithError(err))
			flashutil.SetData(c, flash)
			return c.Redirect().To(c.OriginalURL())
		},
	})
}
