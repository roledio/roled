package project

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/roledio/roled/internal/models"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

// newUpdateProjectSettingsTestCtx builds a context with a system account and returns shared fixtures.
func newUpdateProjectSettingsTestCtx(t *testing.T) (context.Context, string, *entities.Project, *entities.ProjectSetting) {
	t.Helper()
	ctx := context.Background()
	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	projectID := "proj-123"
	project := &entities.Project{
		ID:        projectID,
		AccountID: "acc-123",
		Name:      "Test Project",
		IsActive:  true,
	}
	setting := &entities.ProjectSetting{
		ID:                      "ps-123",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ProjectID:               projectID,
		IsSignupEnabled:         false,
		IsSignupVerifyEmail:     false,
		IsForgotPasswordEnabled: false,
		IsAllowTempEmail:        true,
	}
	return ctx, projectID, project, setting
}

func TestProjectService_UpdateProjectSettings_Success(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

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
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, existingSetting).Return(1, nil)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, []string{rediskeys.ProjectSettingByProjectID(projectID)}).Return(nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsSignupEnabled)
	assert.Equal(t, &roleID, res.DefaultSignupRoleID)
	assert.True(t, res.IsSignupVerifyEmail)
	assert.True(t, res.IsForgotPasswordEnabled)
	assert.False(t, res.IsAllowTempEmail)
}

func TestProjectService_UpdateProjectSettings_SignupDisabled_NoRoleRequired(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, existingSetting).Return(1, nil)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, []string{rediskeys.ProjectSettingByProjectID(projectID)}).Return(nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(false),
		DefaultSignupRoleID:     nil, // no role needed when signup is disabled
		IsSignupVerifyEmail:     pointy.Bool(false),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(true),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.IsSignupEnabled)
	assert.Nil(t, res.DefaultSignupRoleID)
	assert.False(t, res.IsAllowTempEmail)
	assert.False(t, res.IsSignupVerifyEmail)
}

func TestProjectService_UpdateProjectSettings_SignupEnabled_NoRole_ReturnsError(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     nil, // missing when signup is enabled → error
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrDefaultSignupRoleRequired)
}

func TestProjectService_UpdateProjectSettings_SettingsNotFound(t *testing.T) {
	ctx, projectID, project, _ := newUpdateProjectSettingsTestCtx(t)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(nil, nil) // nil = not found

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String("role-123"),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, errors.ErrProjectSettingsNotFound, err)
}

func TestProjectService_UpdateProjectSettings_RoleNotFound(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

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
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, nil) // nil = not found

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrRoleNotFound)
}

func TestProjectService_UpdateProjectSettings_InvalidRoleID_ReturnsError(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

	roleID := "role-belongs-to-other-project"

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	// Role exists in DB but FindByIDAndProjectID returns nil because project_id doesn't match
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrRoleNotFound)
}

func TestProjectService_UpdateProjectSettings_FindSettingsRepoError(t *testing.T) {
	ctx, projectID, project, _ := newUpdateProjectSettingsTestCtx(t)

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
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String("role-123"),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSettings_FindRoleRepoError(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

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
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(nil, dbErr)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSettings_UpdateRepoError(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

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
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, existingSetting).Return(0, dbErr)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestProjectService_UpdateProjectSettings_ZeroRowsAffected(t *testing.T) {
	ctx, projectID, project, existingSetting := newUpdateProjectSettingsTestCtx(t)

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
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(existingSetting, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, roleID, projectID).Return(role, nil)
	mockProjectSettingRepo.EXPECT().Update(ctx, existingSetting).Return(0, nil) // 0 rows affected

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)
	req := &models.UpdateProjectSettingsRequest{
		ProjectID:               projectID,
		IsSignupEnabled:         pointy.Bool(true),
		DefaultSignupRoleID:     pointy.String(roleID),
		IsSignupVerifyEmail:     pointy.Bool(true),
		IsForgotPasswordEnabled: pointy.Bool(true),
		IsAllowTempEmail:        pointy.Bool(false),
	}

	res, err := service.UpdateProjectSettings(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, errors.ErrProjectSettingsNotFound, err)
}
