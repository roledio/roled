package resource

import (
	"context"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgredis "github.com/roledio/roled/auth/pkg/redis"
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
	redisService pkgredis.Service
}

func NewResourceService(registry repositories.Registry, redisService pkgredis.Service) ResourceService {
	return &resourceService{
		registry:     registry,
		redisService: redisService,
	}
}
