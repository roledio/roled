package interfaces

import (
	"context"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
)

type UserRepository interface {
	Count(ctx context.Context, req *models.GetUsersRequest) (int, error)
	FindAll(ctx context.Context, req *models.GetUsersRequest) ([]UserAndRole, error)
	FindByProjectIDAndEmail(ctx context.Context, projectID, email string) (*entities.User, error)
	FindByProjectIDAndExternalUserID(ctx context.Context, projectID, externalUserID string) (*entities.User, error)
	FindByProjectIDAndExternalUserIDJoinRole(ctx context.Context, projectID, externalUserID string) (*UserAndRole, error)
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id string) (*entities.User, error)
	FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.User, error)
	FindByIDAndProjectIDJoinRole(ctx context.Context, userID, projectID string) (*UserAndRole, error)
	SetEmailVerified(ctx context.Context, userID string) (int, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) (int, error)
	Update(ctx context.Context, user *entities.User) (int, error)
	DeleteByID(ctx context.Context, userID string) (int, error)
	DeleteByAccountID(ctx context.Context, accountID string) (int, error)
	DeleteByProjectID(ctx context.Context, projectID string) (int, error)
}

type UserAndRole struct {
	entities.User
	RoleID   string `db:"role_id"`
	RoleName string `db:"role_name"`
}
