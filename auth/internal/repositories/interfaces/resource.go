package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
)

type ResourceRepository interface {
	FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Resource, error)
	Create(ctx context.Context, resources []entities.Resource) (int, error)
	Count(ctx context.Context, req *models.GetResourcesRequest) (int, error)
	FindAll(ctx context.Context, req *models.GetResourcesRequest) ([]entities.Resource, error)
	FindByIDAndProjectID(ctx context.Context, resourceID string, projectID string) (*entities.Resource, error)
	Update(ctx context.Context, resource *entities.Resource) (int, error)
	Delete(ctx context.Context, resource *entities.Resource) (int, error)
}
