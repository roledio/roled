package role

import (
	"context"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgredis "github.com/roledio/roled/auth/pkg/redis"
)

type RoleService interface {
	GetRoles(ctx context.Context, req *models.GetProjectRolesRequest) ([]models.RoleDetails, int, error)
	GetRoleDetails(ctx context.Context, req *models.GetRoleDetailsRequest) (*models.RoleDetails, error)
	CreateRole(ctx context.Context, req *models.CreateRoleRequest) (*models.RoleDetailsAndPermissions, error)
	UpdateRole(ctx context.Context, req *models.UpdateRoleRequest) (*models.RoleDetailsAndPermissions, error)
	DeleteRole(ctx context.Context, req *models.DeleteRoleRequest) error
}

type roleService struct {
	registry     repositories.Registry
	redisService pkgredis.Service
}

func NewRoleService(registry repositories.Registry, redisService pkgredis.Service) RoleService {
	return &roleService{
		registry:     registry,
		redisService: redisService,
	}
}
