package main

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime/debug"

	fibernewrelic "github.com/gofiber/contrib/v3/newrelic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v3"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/handlers/api"
	"github.com/roledio/roled/internal/handlers/web"
	"github.com/roledio/roled/internal/middlewares"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/views"
	pkgconstants "github.com/roledio/roled/pkg/constants"
	"github.com/roledio/roled/pkg/utils/encryptionutil"
	"github.com/roledio/roled/pkg/utils/idutil"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func setupFiberApp(defaultConfig *configs.DefaultConfig, nrapp *newrelic.Application, registry repositories.Registry, redis infra.RedisService, services *Services) *fiber.App {
	buildInfo := models.GetCurrentBuildInfo(defaultConfig)
	engine := html.NewFileSystem(http.FS(views.TemplatesFS), ".html")
	engine.AddFunc("getenv", os.Getenv)

	app := fiber.New(fiber.Config{
		AppName:    fmt.Sprintf("%s v%s-%s", buildInfo.AppName, buildInfo.AppVersion, buildInfo.CommitHash),
		TrustProxy: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: []string{
				"127.0.0.1",
				"::1",
				"10.42.0.0/16", // default pod CIDR k3s
				"10.43.0.0/16", // default service CIDR k3s
			},
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true, // Will return the first IP address if the header contains more than one
		Views:              engine,
		BodyLimit:          10 * 1024 * 1024, // Max body size: 10MB for all requests (uploads will have a separate file size check in the handler)
	})
	if defaultConfig.Newrelic.Enabled && nrapp != nil {
		app.Use(fibernewrelic.New(fibernewrelic.Config{
			Application: nrapp,
		}))
	}
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			log.Panicw("Unexpected error occurred", "panic", e, "stacktrace", string(debug.Stack()))
		},
	}))
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key:    getCookieEncryptionKey(defaultConfig),
		Except: []string{csrf.ConfigDefault.CookieName}, // exclude CSRF cookie
	}))
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return idutil.NanoID()
		},
	}))
	app.Use(middlewares.RequestLogger(buildInfo))
	app.Use(middlewares.CORS(defaultConfig))
	app.Use(middlewares.CSRF(defaultConfig, redis))

	// Fix for GoFiber v3 static middleware breaking changes:
	// 1. Embed prefix stripping: //go:embed "assets" stores files with "assets/" prefix
	//    inside the FS. fs.Sub() strips that prefix so a request for
	//    /assets/static/roled.css resolves to assets/static/roled.css (not assets/assets/static/...).
	// 2. Strict wildcard route "/assets/*" is required in Fiber v3 so sub-paths like
	//    /assets/static/roled.js are matched (Fiber v2 matched these implicitly on app.Use).
	//    Using "/assets/*" (with slash) instead of "/assets*" correctly rejects
	//    malformed paths like /assetsfoo.js and enforces the /assets/<file> structure.
	// 3. When using the FS config field, the first argument (root) must be ""
	//    because the filesystem itself already defines the root.
	assetsSubFS, err := fs.Sub(views.AssetsFS, "assets")
	if err != nil {
		log.Panicw("Failed to create assets sub filesystem", "error", err)
	}
	app.Use("/assets/*", static.New("", static.Config{
		FS: assetsSubFS,
	}))

	// Register uploads static route if using local upload driver
	if defaultConfig.Upload.Driver == constants.UploadDriverLocal {
		// Fiber v3 fix: use "/uploads/*" strict wildcard (with slash) so files like
		// /uploads/logo.png are served, while malformed paths like /uploadsfoo.png
		// are correctly rejected (Fiber v3 no longer matches sub-paths implicitly on Use).
		app.Use("/uploads/*", static.New(defaultConfig.Upload.Local.UploadPath))
	}

	// Enable debug error in non-production environment
	responseutil.DebugError = !defaultConfig.IsEnvProd()

	// Setup handler dependencies
	apiDeps := newApiHandlerDeps(registry, redis, services)
	webDeps := newWebHandlerDeps(registry, redis, services)

	api.NewHandler(app, defaultConfig, apiDeps).SetupRoutes()

	web.NewHandler(app, defaultConfig, webDeps).SetupRoutes()

	return app
}

func getCookieEncryptionKey(defaultConfig *configs.DefaultConfig) string {
	key, err := encryptionutil.DeriveKey([]byte(defaultConfig.EncryptionMasterKey), pkgconstants.KeyPurposeCookie)
	if err != nil {
		log.Error("Failed to derive cookie encryption key: ", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}
