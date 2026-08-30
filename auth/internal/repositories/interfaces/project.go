package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
)

type ProjectRepository interface {
	FindByID(ctx context.Context, id string) (*entities.Project, error)
	FindByIDAndAccountID(ctx context.Context, id string, accountID string) (*entities.Project, error)
	FindSystem(ctx context.Context) (*entities.Project, error)
	FindAll(ctx context.Context, req *models.GetProjectsRequest, accountID string) ([]entities.Project, error)
	Count(ctx context.Context, req *models.GetProjectsRequest, accountID string) (int, error)
	Create(ctx context.Context, project *entities.Project) error
	Update(ctx context.Context, project *entities.Project) (int, error)
	Delete(ctx context.Context, project *entities.Project) (int, error)
}
