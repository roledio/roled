package main

import (
	"fmt"
	"os"
	"time"

	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/queues/handlers"
	"github.com/roledio/roled/internal/queues/publishers"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/accesstoken"
	"github.com/roledio/roled/internal/services/account"
	"github.com/roledio/roled/internal/services/authorize"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/services/member"
	"github.com/roledio/roled/internal/services/project"
	"github.com/roledio/roled/internal/services/upload"
	"github.com/roledio/roled/internal/services/user"
	"github.com/roledio/roled/pkg/utils/cacheutil"
)

type App struct {
	config           *configs.DefaultConfig
	db               *sqlx.DB
	redisService     infra.RedisService
	emailService     infra.EmailService
	authorizeService authorize.AuthorizeService
	projectService   project.ProjectService
	tokenService     accesstoken.AccessTokenService
	accountService   account.AccountService
	userService      user.UserService
	memberService    member.MemberService
	uploadService    upload.UploadService
	queueHandlers    QueueHandlers
	queuePublishers  QueuePublishers
	logger           *fiberzap.LoggerConfig
	app              *fiber.App
}

func NewApp(config *configs.DefaultConfig) (*App, error) {
	buildInfo := models.GetCurrentBuildInfo(config)

	// Setup New Relic
	newrelicService, err := infra.NewNewrelicService(config, buildInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create New Relic service: %w", err)
	}

	// Setup logger
	fiberlogger := setupLogger(config, newrelicService.GetApplication(), buildInfo)

	// Setup database
	db, err := setupDatabase(config, fiberlogger.SetLogger, buildInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	// Setup registry and services
	redisService := infra.NewRedisService(config)
	emailService := infra.NewEmailService(config)
	registry := repositories.NewRegistry(config, db, redisService)

	// Set cache service
	cacheutil.SetService(redisService)

	queuePublishers := QueuePublishers{
		EmailPublisher: publishers.NewEmailPublisher(redisService, constants.QueueEmail, constants.QueueEmailDLQ),
	}

	queueHandlers := QueueHandlers{
		EmailHandler: handlers.NewEmailHandler(config, emailService, redisService),
	}

	services := setupServices(config, registry, queuePublishers, redisService, emailService)

	// Setup Fiber app
	app := setupFiberApp(config, newrelicService.GetApplication(), registry, redisService, services)

	return &App{
		config:           config,
		db:               db,
		redisService:     redisService,
		emailService:     emailService,
		authorizeService: services.AuthorizeService,
		projectService:   services.ProjectService,
		tokenService:     services.TokenService,
		accountService:   services.AccountService,
		userService:      services.UserService,
		memberService:    services.MemberService,
		uploadService:    services.UploadService,
		queueHandlers:    queueHandlers,
		queuePublishers:  queuePublishers,
		logger:           fiberlogger,
		app:              app,
	}, nil
}

func (a *App) Run() error {
	printGeneratedCredentials()
	addr := fmt.Sprintf(":%d", a.config.Port)
	return a.app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true})
}

func (a *App) Shutdown() error {
	return a.app.ShutdownWithTimeout(10 * time.Second)
}

func (a *App) Close() {
	log.Info("Closing application resources...")
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Error("Error closing database: ", err)
		}
	}
	if a.redisService != nil {
		if err := a.redisService.Close(); err != nil {
			log.Error("Error closing Redis service: ", err)
		}
	}
	if a.logger != nil {
		if err := a.logger.Sync(); err != nil {
			log.Error("Error syncing logger: ", err)
		}
	}
	log.Info("Application resources closed.")
}

func printGeneratedCredentials() {
	initialDataGenerated := os.Getenv("INITIAL_DATA_GENERATED")
	if initialDataGenerated != "true" {
		return
	}

	clientID := os.Getenv("GENERATED_CLIENT_ID")
	clientSecret := os.Getenv("GENERATED_CLIENT_SECRET")
	adminEmail := os.Getenv("GENERATED_ADMIN_EMAIL")
	adminPassword := os.Getenv("GENERATED_ADMIN_PASSWORD")
	userEmail := os.Getenv("GENERATED_USER_EMAIL")
	userPassword := os.Getenv("GENERATED_USER_PASSWORD")
	loginURL := os.Getenv("GENERATED_LOGIN_URL")
	redirectURI := os.Getenv("GENERATED_REDIRECT_URI")

	yellow := "\033[1;33m"
	bold := "\033[1m"
	reset := "\033[0m"

	fmt.Println()
	fmt.Println(yellow + "  Welcome to Roled" + reset)
	fmt.Println(yellow + "  Centralized User & Role Management Platform" + reset)
	fmt.Println()
	fmt.Println(bold + "================================================================================" + reset)
	fmt.Println(bold + "                        INITIAL SEED CREDENTIALS                               " + reset)
	fmt.Println(bold + "================================================================================" + reset)
	fmt.Println()
	fmt.Println(yellow + "  Main Client for Roled Console" + reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Client ID", bold, clientID, reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Client Secret", bold, clientSecret, reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Redirect URI", bold, redirectURI, reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Login URL", bold, loginURL, reset)
	fmt.Println()
	fmt.Println(yellow + "  Admin User" + reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Email", bold, adminEmail, reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Password", bold, adminPassword, reset)
	fmt.Println()
	fmt.Println(yellow + "  Standard User" + reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Email", bold, userEmail, reset)
	fmt.Printf("  %-16s : %s%s%s\n", "Password", bold, userPassword, reset)
	fmt.Println()
	fmt.Println(bold + "================================================================================" + reset)
	fmt.Println(yellow + "  Important: Please store these credentials safely for your records." + reset)
	fmt.Println(bold + "================================================================================" + reset)
	fmt.Println()
}
