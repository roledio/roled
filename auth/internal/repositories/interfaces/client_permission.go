package interfaces

import (
	"context"

	"github.com/roledio/roled/internal/entities"
)

type ClientPermissionRepository interface {
	Create(ctx context.Context, clientPermissions []entities.ClientPermission) error
	DeleteByClientID(ctx context.Context, clientID string) (int, error)
	FindByClientID(ctx context.Context, clientID string) ([]entities.ClientPermission, error)
	FindByRoleID(ctx context.Context, roleID string) ([]entities.RolePermission, error)
}
