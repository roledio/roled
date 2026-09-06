package oauthconnection

import (
	"context"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
)

type OAuthConnectionService interface {
	GetOAuthConnections(ctx context.Context, req *models.GetOAuthConnectionsRequest) ([]models.OAuthConnectionDetails, error)
}

type oAuthConnectionService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	redis         infra.RedisService
}

func NewOAuthConnectionService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redis infra.RedisService) OAuthConnectionService {
	return &oAuthConnectionService{
		defaultConfig: defaultConfig,
		registry:      registry,
		redis:         redis,
	}
}
