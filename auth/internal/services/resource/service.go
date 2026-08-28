package resource

import (
	"context"

	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
)

type ResourceService interface {
	GetResources(ctx context.Context, req *models.GetResourcesRequest) ([]models.ResourceDetails, int, error)
	GetResourceDetails(ctx context.Context, req *models.GetResourceDetailsRequest) (*models.ResourceDetails, error)
	CreateResource(ctx context.Context, req *models.CreateResourceRequest) (*models.ResourceDetails, error)
	UpdateResource(ctx context.Context, req *models.UpdateResourceRequest) (*models.ResourceDetails, error)
	DeleteResource(ctx context.Context, req *models.DeleteResourceRequest) error
}

type resourceService struct {
	registry     repositories.Registry
	redisService infra.RedisService
}

func NewResourceService(registry repositories.Registry, redisService infra.RedisService) ResourceService {
	return &resourceService{
		registry:     registry,
		redisService: redisService,
	}
}
