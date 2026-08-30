package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermissions []entities.RolePermission) error
	DeleteByRoleID(ctx context.Context, roleID string) (int, error)
}
