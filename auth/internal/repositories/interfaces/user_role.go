package interfaces

import (
	"context"

	"github.com/roledio/roled/internal/entities"
)

type UserRoleRepository interface {
	Create(ctx context.Context, userRole *entities.UserRole) error
	DeleteByUserID(ctx context.Context, userID string) (int, error)
	FindUserIDsByRoleID(ctx context.Context, roleID string) ([]string, error)
}
