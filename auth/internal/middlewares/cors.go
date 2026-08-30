package middlewares

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/roledio/roled/auth/internal/configs"
)

func CORS(defaultConfig *configs.DefaultConfig) fiber.Handler {
	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			u, err := url.Parse(origin)
			if err != nil {
				log.Error("Parse origin error: ", err)
				return false
			}
			hostname := u.Hostname()
			for _, domain := range defaultConfig.CORS.AllowedDomains {
				domain = strings.TrimSpace(domain)
				if domain != "" && strings.HasSuffix(hostname, domain) {
					return true
				}
			}
			return false
		},
	})
}
