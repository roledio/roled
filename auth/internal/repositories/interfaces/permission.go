package interfaces

import (
	"context"
	"time"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
)

type PermissionRepository interface {
	FindByRoleID(ctx context.Context, roleID string) ([]PermissionResource, error)
	FindByClientID(ctx context.Context, clientID string) ([]PermissionResource, error)
	FindByResourceIDsAndSearch(ctx context.Context, resourceIDs []string, search string) ([]entities.Permission, error)
	Create(ctx context.Context, permissions []entities.Permission) (int, error)
	FindByIDs(ctx context.Context, ids []string) ([]PermissionResource, error)
	DeleteByResourceID(ctx context.Context, resourceID string) (int, error)
	FindAll(ctx context.Context, req *models.GetPermissionsRequest) ([]PermissionResource, error)
	Count(ctx context.Context, req *models.GetPermissionsRequest) (int, error)
	FindByProjectID(ctx context.Context, projectID string, isDefault *bool) ([]entities.Permission, error)
}

// PermissionResource is a struct that represents a permission along with its associated resource information.
type PermissionResource struct {
	ID           string    `db:"id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	ResourceID   string    `db:"resource_id"`
	ResourceName string    `db:"resource_name"`
	ResourceCode string    `db:"resource_code"`
	Code         string    `db:"code"`
	Name         string    `db:"name"`
	Description  *string   `db:"description"`
	IsDefault    bool      `db:"is_default"`
}
