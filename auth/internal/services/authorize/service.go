package authorize

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/shomali11/util/xhashes"
)

type AuthorizeService interface {
	RenderAuthorize(ctx context.Context, req *models.RenderAuthorizeRequest) (*models.RenderAuthorizeResult, error)
	SubmitAuthorize(ctx context.Context, req *models.SubmitAuthorizeRequest) (*models.SubmitAuthorizeResult, error)
	InitiateGoogleOAuth(ctx context.Context, req *models.GoogleOAuthRequest) (string, error)
	HandleGoogleOAuthCallback(ctx context.Context, req *models.GoogleOAuthCallbackRequest) (string, error)
}

type authorizeService struct {
	defaultConfig  *configs.DefaultConfig
	registry       repositories.Registry
	rediService    infra.RedisService
	emailPublisher queues.Publisher
}

func NewAuthorizeService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redisService infra.RedisService,
	emailPublisher queues.Publisher) AuthorizeService {
	return &authorizeService{
		defaultConfig:  defaultConfig,
		registry:       registry,
		rediService:    redisService,
		emailPublisher: emailPublisher,
	}
}

func (s *authorizeService) buildAuthCode(ctx context.Context, user *entities.User, project *entities.Project, req *models.RenderAuthorizeRequest) (string, *entities.AuthCode) {
	now := time.Now()
	expiresAt, err := tparse.AddDuration(now, s.defaultConfig.JWT.AuthCodeExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to calculate auth code expiry time, set default to 2 minutes", "auth_code_expiry_duration", s.defaultConfig.JWT.AuthCodeExpiryDuration)
		expiresAt = now.Add(2 * time.Minute) // default to 2 minutes
	}

	// Generate authorization code
	code := idutil.NanoID(64)

	// Create auth code record
	authCode := &entities.AuthCode{
		ID:                  idutil.NewID(),
		AccountID:           user.AccountID, // Account ID from user to differentiate between system and non-system users
		ProjectID:           project.ID,
		ClientID:            req.ClientID,
		UserID:              &user.ID,
		CodeHash:            xhashes.SHA256(code),
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		RedirectURI:         req.RedirectURI,
		State:               &req.State,
		ExpiresAt:           expiresAt,
	}
	return code, authCode
}
