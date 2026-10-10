package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	queuemocks "github.com/roledio/roled/auth/internal/queues/mocks"
	"github.com/roledio/roled/auth/internal/repositories"
	interfacemocks "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	repositorymocks "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
)

func TestUserService_InviteUser_SystemProjectNotSupported(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	systemProject := &entities.Project{
		ID:        "project-sys",
		AccountID: account.ID,
		IsSystem:  true,
		IsActive:  true,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-sys", account.ID).Return(systemProject, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.InviteUserRequest{
		ProjectID: "project-sys",
		Email:     "user@example.com",
	}

	res, err := service.InviteUser(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, pkgerrors.ErrOperationNotAvailable)
}

func TestUserService_InviteUser_Success(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockUserRepo := interfacemocks.NewMockUserRepository(t)
	mockRoleRepo := interfacemocks.NewMockRoleRepository(t)
	mockUserRoleRepo := interfacemocks.NewMockUserRoleRepository(t)
	mockRedis := redismocks.NewMockService(t)
	mockPublisher := queuemocks.NewMockPublisher(t)

	project := &entities.Project{
		ID:        "project-1",
		AccountID: account.ID,
		Name:      "Test Project",
		IsSystem:  false,
		IsActive:  true,
	}
	role := &entities.Role{
		ID:        "role-1",
		ProjectID: "project-1",
		Name:      "Member",
	}

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(
		func(fn func(repositories.Registry) error) error {
			innerRegistry := repositorymocks.NewMockRegistry(t)
			innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
			innerRegistry.EXPECT().UserRoleRepository().Return(mockUserRoleRepo)
			return fn(innerRegistry)
		},
	)

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "project-1", "user@example.com").Return(nil, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, "role-1", "project-1").Return(role, nil)
	mockUserRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil).Run(func(ctx context.Context, u *entities.User) {
		assert.Equal(t, account.ID, u.AccountID)
		assert.Equal(t, "project-1", u.ProjectID)
		assert.Equal(t, "user@example.com", *u.Email)
		assert.False(t, u.IsActive)
	})
	mockUserRoleRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.UserRole")).Return(nil).Run(func(ctx context.Context, ur *entities.UserRole) {
		assert.Equal(t, "role-1", ur.RoleID)
	})
	mockRedis.EXPECT().SetData(ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)
	mockPublisher.EXPECT().Publish(ctx, mock.Anything).Return(nil)

	cfg := newDefaultConfig()
	cfg.ActivateMemberExpiryDuration = "24h"
	service := NewUserService(cfg, mockRegistry, nil, mockRedis, mockPublisher)

	req := &models.InviteUserRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
		RoleID:    pointy.String("role-1"),
	}

	res, err := service.InviteUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "user@example.com", *res.Email)
}

func TestUserService_InviteUser_EmailAlreadyExists(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)
	mockUserRepo := interfacemocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{
		ID:        "project-1",
		AccountID: account.ID,
		Name:      "Test Project",
		IsSystem:  false,
		IsActive:  true,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "project-1", "user@example.com").Return(&entities.User{ID: "existing-user"}, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.InviteUserRequest{
		ProjectID: "project-1",
		Email:     "user@example.com",
	}

	res, err := service.InviteUser(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrUserEmailAlreadyUsed)
}

func TestUserService_InviteUser_DifferentAccount(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := interfacemocks.NewMockProjectRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Since project belongs to a different account, FindByIDAndAccountID returns nil (not found for this account)
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-of-other-account", account.ID).Return(nil, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.InviteUserRequest{
		ProjectID: "project-of-other-account",
		Email:     "user@example.com",
	}

	res, err := service.InviteUser(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, errors.ErrProjectNotFound)
}
