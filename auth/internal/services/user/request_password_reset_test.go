package user

import (
	"context"
	"testing"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	queuemocks "github.com/roledio/roled/internal/mocks/queues"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/roledio/roled/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"
)

func TestUserService_RequestPasswordReset_Success_WithoutRedirectURI(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockPublisher := queuemocks.NewMockPublisher(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}
	user := &entities.User{
		ID:          "user-1",
		ProjectID:   "project-1",
		Email:       pointy.String("user@example.com"),
		DisplayName: "Test User",
		IsActive:    true,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(user, nil)
	mockRedis.EXPECT().SetData(ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	mockPublisher.EXPECT().Publish(ctx, mock.Anything).Return(nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, mockRedis, mockPublisher)
	req := &models.RequestPasswordResetRequest{
		ProjectID: "project-1",
		UserID:    "user-1",
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.NoError(t, err)
}

func TestUserService_RequestPasswordReset_Success_WithRedirectURI(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockPublisher := queuemocks.NewMockPublisher(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}
	user := &entities.User{
		ID:          "user-1",
		ProjectID:   "project-1",
		Email:       pointy.String("user@example.com"),
		DisplayName: "Test User",
		IsActive:    true,
	}
	redirectURI := &entities.RedirectURI{
		ProjectID:   "project-1",
		RedirectURI: "https://example.com/callback",
		LoginURL:    pointy.String("https://example.com/login"),
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(user, nil)
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "project-1", "https://example.com/callback").Return(redirectURI, nil)
	mockRedis.EXPECT().SetData(ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	mockPublisher.EXPECT().Publish(ctx, mock.Anything).Return(nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, mockRedis, mockPublisher)
	req := &models.RequestPasswordResetRequest{
		ProjectID:   "project-1",
		UserID:      "user-1",
		RedirectURI: pointy.String("https://example.com/callback"),
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.NoError(t, err)
}

func TestUserService_RequestPasswordReset_UserNotFound(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(nil, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.RequestPasswordResetRequest{
		ProjectID: "project-1",
		UserID:    "user-1",
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.Equal(t, errors.ErrUserNotFound, err)
}

func TestUserService_RequestPasswordReset_UserNoEmail(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}
	user := &entities.User{
		ID:          "user-1",
		ProjectID:   "project-1",
		Email:       nil,
		DisplayName: "Test User",
		IsActive:    true,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(user, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.RequestPasswordResetRequest{
		ProjectID: "project-1",
		UserID:    "user-1",
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.Equal(t, errors.ErrUserHasNoEmail, err)
}

func TestUserService_RequestPasswordReset_UserInactive(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}
	user := &entities.User{
		ID:          "user-1",
		ProjectID:   "project-1",
		Email:       pointy.String("user@example.com"),
		DisplayName: "Test User",
		IsActive:    false,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(user, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.RequestPasswordResetRequest{
		ProjectID: "project-1",
		UserID:    "user-1",
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.Equal(t, errors.ErrUserNotActive, err)
}

func TestUserService_RequestPasswordReset_RedirectURINotFound(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, Name: "Test Project"}
	user := &entities.User{
		ID:          "user-1",
		ProjectID:   "project-1",
		Email:       pointy.String("user@example.com"),
		DisplayName: "Test User",
		IsActive:    true,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "project-1").Return(user, nil)
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "project-1", "https://invalid.com/callback").Return(nil, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.RequestPasswordResetRequest{
		ProjectID:   "project-1",
		UserID:      "user-1",
		RedirectURI: pointy.String("https://invalid.com/callback"),
	}

	err := service.RequestPasswordReset(ctx, req)
	assert.Equal(t, errors.ErrRedirectURINotFound, err)
}
