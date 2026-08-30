package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
)

type RoleRepository interface {
	Count(ctx context.Context, req *models.GetProjectRolesRequest) (int, error)
	FindAll(ctx context.Context, req *models.GetProjectRolesRequest) ([]entities.Role, error)
	FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Role, error)
	FindByUserID(ctx context.Context, userID string) (*entities.Role, error)
	Create(ctx context.Context, role *entities.Role) error
	FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Role, error)
	Update(ctx context.Context, role *entities.Role) (int, error)
	DeleteByID(ctx context.Context, id string) (int, error)
}
