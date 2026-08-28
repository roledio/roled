package permission

import (
	"context"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
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
