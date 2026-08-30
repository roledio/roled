package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

func TestUserService_GetUsers_Success(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	req := &models.GetUsersRequest{ProjectID: "project-1"}
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, IsSystem: false}
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)

	now := time.Now().UTC()
	userRows := []interfaces.UserAndRole{
		{
			User: entities.User{
				ID:              "user-1",
				CreatedAt:       now,
				UpdatedAt:       now,
				Email:           pointy.String("user@example.com"),
				ExternalUserID:  pointy.String("external-123"),
				DisplayName:     "User One",
				AvatarURL:       pointy.String("http://example.com/avatar.png"),
				IsActive:        true,
				EmailVerifiedAt: pointy.Pointer(now),
			},
			RoleID:   "role-1",
			RoleName: "Admin",
		},
	}

	mockUserRepo.EXPECT().Count(ctx, req).Return(1, nil)
	mockUserRepo.EXPECT().FindAll(ctx, req).Return(userRows, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	users, count, err := service.GetUsers(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, users, 1)
	assert.Equal(t, "user-1", users[0].ID)
	assert.Equal(t, "user@example.com", *users[0].Email)
	assert.Equal(t, "external-123", *users[0].ExternalUserID)
	assert.Equal(t, "User One", users[0].DisplayName)
	assert.Equal(t, "http://example.com/avatar.png", *users[0].AvatarURL)
	assert.True(t, users[0].IsActive)
	assert.True(t, users[0].IsEmailVerified)
	assert.Equal(t, "role-1", users[0].RoleID)
	assert.Equal(t, "Admin", users[0].RoleName)
}

func TestUserService_GetUsers_NoUsers(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	req := &models.GetUsersRequest{ProjectID: "project-1"}
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, IsSystem: false}
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().Count(ctx, req).Return(0, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	users, count, err := service.GetUsers(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, users)
}
