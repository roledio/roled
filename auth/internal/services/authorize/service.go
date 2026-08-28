package authorize

import (
	"context"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
)

type AuthorizeService interface {
	RenderAuthorize(ctx context.Context, req *models.RenderAuthorizeRequest) (*models.RenderAuthorizeResult, error)
	SubmitAuthorize(ctx context.Context, req *models.SubmitAuthorizeRequest) (*models.SubmitAuthorizeResult, error)
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
