package permission

import (
	"context"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
)

type PermissionService interface {
	GetPermissions(ctx context.Context, req *models.GetPermissionsRequest) ([]models.PermissionDetails, int, error)
}

type permissionService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
}

func NewPermissionService(defaultConfig *configs.DefaultConfig, registry repositories.Registry) PermissionService {
	return &permissionService{
		defaultConfig: defaultConfig,
		registry:      registry,
	}
}
