package client

import (
	"context"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
)

type ClientService interface {
	GetClients(ctx context.Context, req *models.GetClientsRequest) ([]models.ClientDetails, int, error)
	GetClientDetails(ctx context.Context, req *models.GetClientDetailsRequest) (*models.ClientDetails, error)
	CreateClient(ctx context.Context, req *models.CreateClientRequest) (*models.ClientDetailsAndPermissions, error)
	UpdateClient(ctx context.Context, req *models.UpdateClientRequest) (*models.ClientDetailsAndPermissions, error)
	DeleteClient(ctx context.Context, req *models.DeleteClientRequest) error
}

type clientService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	redisService  infra.RedisService
}

func NewClientService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redisService infra.RedisService) ClientService {
	return &clientService{
		defaultConfig: defaultConfig,
		registry:      registry,
		redisService:  redisService,
	}
}
