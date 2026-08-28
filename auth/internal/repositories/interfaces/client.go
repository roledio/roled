package interfaces

import (
	"context"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
)

type ClientRepository interface {
	FindByID(ctx context.Context, id string) (*entities.Client, error)
	FindByProjectIDAndIsDefault(ctx context.Context, projectID string, isDefault bool) (*entities.Client, error)
	Create(ctx context.Context, client *entities.Client) error
	DeleteByProjectID(ctx context.Context, projectID string) (int, error)
	Count(ctx context.Context, req *models.GetClientsRequest) (int, error)
	FindAll(ctx context.Context, req *models.GetClientsRequest) ([]entities.Client, error)
	FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Client, error)
	Update(ctx context.Context, client *entities.Client) (int, error)
	Delete(ctx context.Context, client *entities.Client) (int, error)
}
