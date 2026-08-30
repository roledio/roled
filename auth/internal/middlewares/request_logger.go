package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/roledio/roled/auth/internal/constants"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
)

func RequestLogger(buildInfo pkgmodels.BuildInfo) fiber.Handler {
	return func(c fiber.Ctx) error {
		mapRequestLog := map[string]any{
			constants.RequestLogRequestID:  requestid.FromContext(c),
			constants.RequestLogPath:       c.Path(),
			constants.RequestLogMethod:     c.Method(),
			constants.RequestLogIP:         getRealIP(c),
			constants.RequestLogHost:       c.Hostname(),
			constants.RequestLogOrigin:     c.Get("Origin"),
			constants.RequestLogReferer:    c.Get("Referer"),
			constants.RequestLogUserAgent:  c.Get("User-Agent"),
			constants.RequestLogEnv:        buildInfo.Env,
			constants.RequestLogAppName:    buildInfo.AppName,
			constants.RequestLogAppVersion: buildInfo.AppVersion,
			constants.RequestLogCommitHash: buildInfo.CommitHash,
		}
		ctx := c.Context()
		for _, key := range constants.RequestLoggerKeys {
			if v, ok := mapRequestLog[key]; ok {
				// Forced to use string context key since the fiberzap's LoggerConfig.ExtraKeys cannot accept types other than string.
				// The request id will not be printed on the log if the context key is not using string.
				//nolint:staticcheck
				ctx = context.WithValue(ctx, key, v)
			}
		}
		c.SetContext(ctx)
		return c.Next()
	}
}

func getRealIP(c fiber.Ctx) string {
	// Check if the request if coming from Cloudflare by checking the CF-Connecting-IP header,
	// if exists, use it as the real IP instead of the X-Forwarded-For header.
	// This is because Cloudflare will overwrite the X-Forwarded-For header with the Cloudflare IPs,
	// and the real client IP will be in the CF-Connecting-IP header.
	if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// 2. Fallback to Fiber IP (respect TrustProxyConfig)
	return c.IP()
}
