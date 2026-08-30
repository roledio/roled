package project

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	apperrors "github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

// newUpdateSignupRoleTestCtx returns a context with an active system account and shared test fixtures.
func newUpdateSignupRoleTestCtx(t *testing.T) (context.Context, string, *entities.Project, *entities.ProjectSetting) {
	t.Helper()
	ctx := context.Background()
	ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	})
	projectID := "proj-123"
	project := &entities.Project{
		ID:        projectID,
		AccountID: "acc-123",
		Name:      "Test Project",
		IsActive:  true,
	}
	setting := &entities.ProjectSetting{
		ID:              "ps-123",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ProjectID:       projectID,
		IsSignupEnabled: true, // signup is enabled by default in tests
	}
	return ctx, projectID, project, setting
}

func TestProjectService_UpdateProjectSignupRole_Success(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "role-123"
	role := &entities.Role{ID: roleID, ProjectID: projectID, Code: "default-role", Name: "Default Role"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, setting).Return(1, nil)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, []string{rediskeys.ProjectSettingByProjectID(projectID)}).Return(nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, roleID, res.RoleID)
	assert.Equal(t, role.Name, res.RoleName)
}

func TestProjectService_UpdateProjectSignupRole_SignupDisabled_ReturnsError(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)
	setting.IsSignupEnabled = false // signup is disabled

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    "role-123",
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, apperrors.ErrSignupMustBeEnabled, err)
}

func TestProjectService_UpdateProjectSignupRole_SettingsNotFound(t *testing.T) {
	ctx, projectID, project, _ := newUpdateSignupRoleTestCtx(t)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(nil, nil) // not found

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    "role-123",
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, apperrors.ErrProjectSettingsNotFound, err)
}

func TestProjectService_UpdateProjectSignupRole_RoleNotFound(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "unknown-role"

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, nil) // not found

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, apperrors.ErrRoleNotFound, err)
}

func TestProjectService_UpdateProjectSignupRole_RoleDoesNotBelongToProject(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "role-other-project"

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	// FindByIDAndProjectID returns nil because the role doesn't belong to this project
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, apperrors.ErrRoleNotFound, err)
}

func TestProjectService_UpdateProjectSignupRole_FindSettingsRepoError(t *testing.T) {
	ctx, projectID, project, _ := newUpdateSignupRoleTestCtx(t)

	dbErr := fmt.Errorf("db connection failed")

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(nil, dbErr)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    "role-123",
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSignupRole_FindRoleRepoError(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "role-123"
	dbErr := fmt.Errorf("db connection failed")

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, dbErr)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSignupRole_UpdateRepoError(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "role-123"
	role := &entities.Role{ID: roleID, ProjectID: projectID, Code: "default-user"}
	dbErr := fmt.Errorf("db write failed")

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, setting).Return(0, dbErr)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSignupRole_ZeroRowsAffected(t *testing.T) {
	ctx, projectID, project, setting := newUpdateSignupRoleTestCtx(t)

	roleID := "role-123"
	role := &entities.Role{ID: roleID, ProjectID: projectID, Code: "default-user"}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, setting).Return(0, nil) // 0 rows affected

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSignupRoleRequest{
		ProjectID: projectID,
		RoleID:    roleID,
	}

	res, err := service.UpdateProjectSignupRole(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, apperrors.ErrProjectSettingsNotFound, err)
}
