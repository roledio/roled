package main

import (
	"fmt"
	"io"
	"os"

	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/queues"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/accesstoken"
	"github.com/roledio/roled/auth/internal/services/account"
	"github.com/roledio/roled/auth/internal/services/authorize"
	"github.com/roledio/roled/auth/internal/services/client"
	"github.com/roledio/roled/auth/internal/services/infra"
	"github.com/roledio/roled/auth/internal/services/member"
	"github.com/roledio/roled/auth/internal/services/permission"
	"github.com/roledio/roled/auth/internal/services/project"
	"github.com/roledio/roled/auth/internal/services/resource"
	"github.com/roledio/roled/auth/internal/services/role"
	"github.com/roledio/roled/auth/internal/services/upload"
	"github.com/roledio/roled/auth/internal/services/user"
	_ "github.com/roledio/roled/auth/migrations"
	"github.com/roledio/roled/auth/pkg/databases"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
	sqldblogger "github.com/simukti/sqldb-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type QueuePublishers struct {
	EmailPublisher queues.Publisher
}

type QueueHandlers struct {
	EmailHandler queues.Handler
}

type Services struct {
	AuthorizeService  authorize.AuthorizeService
	ProjectService    project.ProjectService
	TokenService      accesstoken.AccessTokenService
	AccountService    account.AccountService
	UserService       user.UserService
	MemberService     member.MemberService
	UploadService     upload.UploadService
	ClientService     client.ClientService
	ResourceService   resource.ResourceService
	RoleService       role.RoleService
	PermissionService permission.PermissionService
}

func setupServices(config *configs.DefaultConfig, registry repositories.Registry, publishers QueuePublishers, redis infra.RedisService, _ infra.EmailService) *Services {
	uploadService := upload.NewUploadService(config)
	return &Services{
		AuthorizeService:  authorize.NewAuthorizeService(config, registry, redis, publishers.EmailPublisher),
		ProjectService:    project.NewProjectService(config, registry, uploadService, redis),
		TokenService:      accesstoken.NewAccessTokenService(config, registry, redis),
		AccountService:    account.NewAccountService(registry, redis),
		UserService:       user.NewUserService(config, registry, uploadService, redis, publishers.EmailPublisher),
		MemberService:     member.NewMemberService(config, registry, publishers.EmailPublisher, redis),
		UploadService:     uploadService,
		ClientService:     client.NewClientService(config, registry, redis),
		ResourceService:   resource.NewResourceService(registry, redis),
		RoleService:       role.NewRoleService(registry, redis),
		PermissionService: permission.NewPermissionService(config, registry),
	}
}

func setupLogger(defaultConfig *configs.DefaultConfig, nrapp *newrelic.Application, buildInfo pkgmodels.BuildInfo) *fiberzap.LoggerConfig {
	// Setup lumberjack for log rotation
	lumberjack := &lumberjack.Logger{
		Filename:  fmt.Sprintf("logs/%s.log", buildInfo.ProjectName),
		MaxAge:    14,
		LocalTime: true,
	}

	// Use stderr and lumberjack as the default log writers
	writer := io.MultiWriter(os.Stderr, lumberjack)

	// If newrelic is enabled, use newrelic log writer in addition to lumberjack
	if defaultConfig.Newrelic.Enabled && nrapp != nil {
		nrlogwriter := logWriter.New(os.Stderr, nrapp)
		nrlogwriter.DebugLogging(true)
		writer = io.MultiWriter(nrlogwriter, lumberjack)
	}

	level := zap.DebugLevel
	if defaultConfig.IsEnvProd() {
		level = zap.InfoLevel
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(writer), level)
	zaplogger := zap.New(core)
	fiberlogger := fiberzap.NewLogger(fiberzap.LoggerConfig{
		SetLogger: zaplogger,
		ExtraKeys: constants.RequestLoggerKeys, // Will be used when calling: log.WithContext(ctx)...
	})

	log.SetLogger(fiberlogger)

	return fiberlogger
}

func setupDatabase(defaultConfig *configs.DefaultConfig, zaplogger *zap.Logger, buildInfo pkgmodels.BuildInfo) (*sqlx.DB, error) {
	config := databases.Config{
		Driver:          databases.DriverMySQL,
		Host:            defaultConfig.DB.Host,
		Port:            defaultConfig.DB.Port,
		Username:        defaultConfig.DB.Username,
		Password:        defaultConfig.DB.Password,
		Name:            defaultConfig.DB.Name,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ApplicationName: buildInfo.ProjectName,
		Newrelic:        defaultConfig.Newrelic.Enabled,
	}
	db, err := databases.Open(config)
	if err != nil {
		return nil, fmt.Errorf("connect to database error: %w", err)
	}
	zapAdapter := databases.NewZapAdapter(zaplogger, constants.RequestLoggerKeys)
	db = sqldblogger.OpenDriver(config.GetDSN(), db.Driver(), zapAdapter,
		sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithPreparerLevel(sqldblogger.LevelTrace),
		sqldblogger.WithTimeFormat(sqldblogger.TimeFormatRFC3339),
	)
	err = databases.Migrate(db, databases.DialectMySQL, "./migrations")
	if err != nil {
		return nil, fmt.Errorf("migrate database error: %w", err)
	}
	return sqlx.NewDb(db, config.GetDriver()), nil
}
