package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
)

func TestUserService_CreateUser_SuccessWithEmailAndAvatar(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockUserRoleRepo := repositorymocks.NewMockUserRoleRepository(t)
	mockUpload := servicemocks.NewMockUploadService(t)

	project := &entities.Project{
		ID:        "project-1",
		AccountID: account.ID,
		IsSystem:  false,
	}
	role := &entities.Role{
		ID:   "role-1",
		Name: "Admin",
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
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "project-1", "test@example.com").Return(nil, nil)
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, "role-1", "project-1").Return(role, nil)
	mockUserRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil).Run(func(ctx context.Context, user *entities.User) {
		assert.Equal(t, "account-1", user.AccountID)
		assert.Equal(t, "project-1", user.ProjectID)
		assert.Equal(t, "test@example.com", *user.Email)
		assert.Equal(t, "Admin user", user.DisplayName)
		assert.NotNil(t, user.PasswordHash)
		assert.NotNil(t, user.AvatarURL)
		assert.Equal(t, "http://localhost/uploads/avatar.png", *user.AvatarURL)
	})
	mockUserRoleRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.UserRole")).Return(nil).Run(func(ctx context.Context, userRole *entities.UserRole) {
		assert.Equal(t, "role-1", userRole.RoleID)
	})
	mockUpload.EXPECT().Move(ctx, "tmp/avatar.png", "avatar.png").Return(nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, mockUpload, nil, nil)
	avatarURL := "http://localhost/uploads/tmp/avatar.png"
	req := &models.CreateUserRequest{
		ProjectID:   "project-1",
		DisplayName: "Admin user",
		Email:       "test@example.com",
		Password:    "securePass123",
		AvatarURL:   avatarURL,
		RoleID:      "role-1",
	}

	res, err := service.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "test@example.com", *res.Email)
	assert.Equal(t, "Admin", res.RoleName)
	assert.Equal(t, "role-1", res.RoleID)
	assert.NotNil(t, res.AvatarURL)
	assert.Equal(t, "http://localhost/uploads/avatar.png", *res.AvatarURL)
}

func TestUserService_CreateUser_EmailAlreadyUsed(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "account-1",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{
		ID:        "project-1",
		AccountID: account.ID,
		IsSystem:  false,
	}

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "project-1", "test@example.com").Return(&entities.User{ID: "existing-user"}, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	req := &models.CreateUserRequest{
		ProjectID:   "project-1",
		DisplayName: "Admin user",
		Email:       "test@example.com",
		Password:    "securePass123",
	}

	res, err := service.CreateUser(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, errors.ErrUserEmailAlreadyUsed, err)
}
