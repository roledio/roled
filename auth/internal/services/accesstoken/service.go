package accesstoken

import (
	"context"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
)

const tokenTypeBearer = "bearer"

type AccessTokenService interface {
	ExchangeToken(ctx context.Context, req *models.ExchangeTokenRequest) (*models.ExchangeTokenResponse, error)
	GetCurrentAccessToken(ctx context.Context) (*models.AccessTokenDetails, error)
	RevokeCurrentToken(ctx context.Context, req *models.RevokeCurrentTokenRequest) error
}

type accessTokenService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	redisService  infra.RedisService
}

func NewAccessTokenService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redisService infra.RedisService) AccessTokenService {
	return &accessTokenService{
		defaultConfig: defaultConfig,
		registry:      registry,
		redisService:  redisService,
	}
}
