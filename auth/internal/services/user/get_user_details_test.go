package user

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/stretchr/testify/assert"
	"go.openly.dev/pointy"
)

func TestUserService_GetUserDetails_Success(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	req := &models.GetUserDetailsRequest{ProjectID: "project-1", UserID: "user-1"}
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, IsSystem: false}
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)

	now := time.Now().UTC()
	userRow := &interfaces.UserAndRole{
		User: entities.User{
			ID:              "user-1",
			CreatedAt:       now,
			UpdatedAt:       now,
			Email:           pointy.String("detail@example.com"),
			DisplayName:     "Detail User",
			AvatarURL:       pointy.String("http://example.com/avatar.png"),
			IsActive:        false,
			EmailVerifiedAt: nil,
		},
		RoleID:   "role-2",
		RoleName: "Member",
	}
	mockUserRepo.EXPECT().FindByIDAndProjectIDJoinRole(ctx, "user-1", "project-1").Return(userRow, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	userDetails, err := service.GetUserDetails(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, userDetails)
	assert.Equal(t, "user-1", userDetails.ID)
	assert.Equal(t, "detail@example.com", *userDetails.Email)
	assert.Equal(t, "Detail User", userDetails.DisplayName)
	assert.Equal(t, "http://example.com/avatar.png", *userDetails.AvatarURL)
	assert.False(t, userDetails.IsActive)
	assert.False(t, userDetails.IsEmailVerified)
	assert.Equal(t, "role-2", userDetails.RoleID)
	assert.Equal(t, "Member", userDetails.RoleName)
}

func TestUserService_GetUserDetails_NotFound(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{ID: "account-1", IsSystem: false}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	req := &models.GetUserDetailsRequest{ProjectID: "project-1", UserID: "user-1"}
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	project := &entities.Project{ID: "project-1", AccountID: account.ID, IsSystem: false}
	mockProjectRepo.EXPECT().FindByIDAndAccountID(ctx, "project-1", account.ID).Return(project, nil)
	mockUserRepo.EXPECT().FindByIDAndProjectIDJoinRole(ctx, "user-1", "project-1").Return(nil, nil)

	service := NewUserService(newDefaultConfig(), mockRegistry, nil, nil, nil)
	userDetails, err := service.GetUserDetails(ctx, req)

	assert.Nil(t, userDetails)
	assert.Equal(t, errors.ErrUserNotFound, err)
}
