package role

import (
	"context"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
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
	redisService infra.RedisService
}

func NewRoleService(registry repositories.Registry, redisService infra.RedisService) RoleService {
	return &roleService{
		registry:     registry,
		redisService: redisService,
	}
}
