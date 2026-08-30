package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
)

func TestUserService_UpdateUser_ChangeAvatar_Success(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "acc-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRoleRepo := repositorymocks.NewMockRoleRepository(t)
	mockUserRoleRepo := repositorymocks.NewMockUserRoleRepository(t)
	mockUpload := servicemocks.NewMockUploadService(t)

	project := &entities.Project{ID: "proj-1", AccountID: account.ID, IsSystem: false}

	existingTime := time.Now().UTC()
	existing := &entities.User{
		ID:              "user-1",
		CreatedAt:       existingTime,
		UpdatedAt:       existingTime,
		Email:           pointy.String("old@example.com"),
		DisplayName:     "Old Name",
		AvatarURL:       pointy.String("http://localhost/uploads/old.png"),
		IsActive:        true,
		EmailVerifiedAt: &existingTime,
	}

	// Registry and repo expectations before tx
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().RoleRepository().Return(mockRoleRepo)

	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "proj-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "proj-1").Return(existing, nil)
	// Email uniqueness check for new email
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "proj-1", "new@example.com").Return(nil, nil)
	// Role validation
	mockRoleRepo.EXPECT().FindByIDAndProjectID(ctx, "role-2", "proj-1").Return(&entities.Role{ID: "role-2", Name: "Member"}, nil)

	// Tx expectations: return inner registry that returns mock repos
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(
		func(fn func(repositories.Registry) error) error {
			inner := repositorymocks.NewMockRegistry(t)
			inner.EXPECT().UserRepository().Return(mockUserRepo)
			inner.EXPECT().UserRoleRepository().Return(mockUserRoleRepo)
			return fn(inner)
		},
	)

	// Inside tx: expect update, delete old role, create new role
	mockUserRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.User")).Return(1, nil)
	mockUserRoleRepo.EXPECT().DeleteByUserID(mock.Anything, "user-1").Return(1, nil)
	mockUserRoleRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*entities.UserRole")).Return(nil)

	// Upload operations: move tmp -> new, delete old file
	mockUpload.EXPECT().Move(ctx, "tmp/new.png", "new.png").Return(nil)
	mockUpload.EXPECT().Delete(ctx, "old.png").Return(nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, mockUpload, nil, nil)

	avatarURL := "http://localhost/uploads/tmp/new.png"
	isActive := true
	req := &models.UpdateUserRequest{
		ProjectID:   "proj-1",
		UserID:      "user-1",
		DisplayName: "New Name",
		AvatarURL:   avatarURL,
		Email:       "new@example.com",
		Password:    "",
		IsActive:    &isActive,
		RoleID:      "role-2",
	}

	res, err := service.UpdateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "user-1", res.ID)
	assert.Equal(t, "New Name", res.DisplayName)
	assert.Equal(t, "role-2", res.RoleID)
	assert.Equal(t, "Member", res.RoleName)
	assert.NotNil(t, res.AvatarURL)
	assert.Equal(t, "http://localhost/uploads/new.png", *res.AvatarURL)
}

func TestUserService_UpdateUser_EmailAlreadyUsed(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "acc-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "proj-1", AccountID: account.ID, IsSystem: false}
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "proj-1", account.ID).Return(project, nil)

	existing := &entities.User{ID: "user-1", Email: pointy.String("old@example.com")}
	mockUserRepo.EXPECT().FindByIDAndProjectID(ctx, "user-1", "proj-1").Return(existing, nil)
	// Email uniqueness check returns existing user -> error
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "proj-1", "taken@example.com").Return(&entities.User{ID: "other"}, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	isActive := true
	req := &models.UpdateUserRequest{
		ProjectID:   "proj-1",
		UserID:      "user-1",
		DisplayName: "Name",
		Email:       "taken@example.com",
		IsActive:    &isActive,
	}

	res, err := service.UpdateUser(ctx, req)

	assert.Nil(t, res)
	assert.Equal(t, errors.ErrUserEmailAlreadyUsed, err)
}
