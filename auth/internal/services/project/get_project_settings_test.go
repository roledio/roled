package project

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
)

func TestProjectService_GetProjectSettings_Success(t *testing.T) {
	ctx := context.Background()

	// Setup system account in context
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
		IsSignupEnabled:         true,
		IsSignupVerifyEmail:     true,
		IsForgotPasswordEnabled: true,
		IsAllowTempEmail:        false,
	}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)

	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil)
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).Return(setting, nil)

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)

	req := &models.GetProjectSettingsRequest{
		ProjectID: projectID,
	}

	res, err := service.GetProjectSettings(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsSignupEnabled)
	assert.True(t, res.IsSignupVerifyEmail)
	assert.True(t, res.IsForgotPasswordEnabled)
	assert.False(t, res.IsAllowTempEmail)
}

func TestProjectService_GetProjectSettings_SingleflightDeduplication(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "acc-123",
		IsSystem: true,
		IsActive: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	projectID := "proj-singleflight-123"
	project := &entities.Project{
		ID:        projectID,
		AccountID: "acc-123",
		Name:      "Test Project Singleflight",
		IsActive:  true,
	}

	setting := &entities.ProjectSetting{
		ID:                      "ps-123",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ProjectID:               projectID,
		IsSignupEnabled:         true,
		IsSignupVerifyEmail:     true,
		IsForgotPasswordEnabled: true,
		IsAllowTempEmail:        false,
	}

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	// Since calls are singleflighted, underlying repository methods are called ONLY ONCE even with 10 concurrent callers.
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo).Once()
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo).Once()
	mockProjectRepo.EXPECT().FindByID(ctx, projectID).Return(project, nil).Once()
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, projectID).RunAndReturn(func(ctx context.Context, pID string) (*entities.ProjectSetting, error) {
		time.Sleep(50 * time.Millisecond) // Simulate DB delay so concurrent calls overlap
		return setting, nil
	}).Once()

	service := NewProjectService(&configs.DefaultConfig{}, mockRegistry, nil, mockRedisService)

	const numGoroutines = 10
	results := make([]*models.ProjectSettings, numGoroutines)
	errs := make([]error, numGoroutines)

	startGate := make(chan struct{})
	done := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			<-startGate
			req := &models.GetProjectSettingsRequest{ProjectID: projectID}
			res, err := service.GetProjectSettings(ctx, req)
			results[idx] = res
			errs[idx] = err
			done <- struct{}{}
		}(i)
	}

	close(startGate)
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errs[i])
		assert.NotNil(t, results[i])
		assert.True(t, results[i].IsSignupEnabled)
	}
}
