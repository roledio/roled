package member

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	pkgerrors "github.com/roledio/roled/pkg/errors"

	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
)

func TestMemberService_DeleteMember_SystemProjectNotFound(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project not found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(nil, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.Msg, err.Error())
}

func TestMemberService_DeleteMember_SystemProjectFindError(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project find error
	dbErr := assert.AnError
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(nil, dbErr)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestMemberService_DeleteMember_AccessTokenNotInContext(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccessTokenNotFound, err)
}

func TestMemberService_DeleteMember_ProjectIDMismatch(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "wrong-project-id", // Different from system project
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrOperationNotAvailable, err)
}

func TestMemberService_DeleteMember_AccountNotInContext(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccountNotFound, err)
}

func TestMemberService_DeleteMember_ClientJWT_SystemAccount_Success(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // For validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock delete operations
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByUserID(ctx, "target-user-id").Return(0, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestMemberService_DeleteMember_ClientJWT_SystemAccount_LastAdmin(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   true,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (only 1 admin, cannot delete)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(1, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCannotDeleteLastAdmin, err)
}

func TestMemberService_DeleteMember_ClientJWT_NonSystemAccount_Success(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "test-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // For validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "test-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock delete operations
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByUserID(ctx, "target-user-id").Return(0, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestMemberService_DeleteMember_ClientJWT_NonSystemAccount_DifferentAccount(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account", // Different account
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_DeleteMember_ClientJWT_MemberNotFound(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member not found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(nil, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_DeleteMember_UserJWT_SystemAccount_Success(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // For validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock delete operations
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByUserID(ctx, "target-user-id").Return(0, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestMemberService_DeleteMember_UserJWT_SystemAccount_DeleteSelf(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "current-user-id", // Same as current user
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCannotDeleteSelf, err)
}

func TestMemberService_DeleteMember_UserJWT_NonSystemAccount_Success(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	currentMember := &entities.Member{
		ID:        "current-member-id",
		UserID:    "current-user-id",
		AccountID: "test-account",
		IsAdmin:   true,
	}

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "test-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo).Times(1) // Once for user JWT handling

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock current member lookup
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "current-user-id").Return(currentMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "test-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock delete operations
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByUserID(ctx, "target-user-id").Return(0, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestMemberService_DeleteMember_UserJWT_NonSystemAccount_DeleteSelf(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	currentMember := &entities.Member{
		ID:        "current-member-id",
		UserID:    "current-user-id",
		AccountID: "test-account",
		IsAdmin:   true,
	}

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "current-user-id", // Same as current user
		AccountID: "test-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Once for user JWT handling

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock current member lookup
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "current-user-id").Return(currentMember, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCannotDeleteSelf, err)
}

func TestMemberService_DeleteMember_UserJWT_NonSystemAccount_NonAdminUser(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	currentMember := &entities.Member{
		ID:        "current-member-id",
		UserID:    "current-user-id",
		AccountID: "test-account",
		IsAdmin:   false, // Not admin
	}

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "test-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Once for user JWT handling

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock current member lookup
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "current-user-id").Return(currentMember, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNonAdminDeleteMember, err)
}

func TestMemberService_DeleteMember_UserJWT_NonSystemAccount_DifferentAccount(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    pointy.String("current-user-id"), // User JWT
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account", // Different account
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Only once for target member

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestMemberService_DeleteMember_TransactionFailure_MemberDeleteError(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Once for validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		// UserRepository and AccessTokenRepository are not called because member delete fails
		return fn(innerRegistry)
	})

	// Mock delete operations - member delete fails
	dbErr := assert.AnError
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(0, dbErr)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestMemberService_DeleteMember_TransactionFailure_UserDeleteError(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Once for validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		// AccessTokenRepository is not called because user delete fails
		return fn(innerRegistry)
	})

	// Mock delete operations - member delete succeeds, user delete fails
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	dbErr := assert.AnError
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(0, dbErr)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}

func TestMemberService_DeleteMember_TransactionFailure_AccessTokenDeleteError(t *testing.T) {
	ctx := context.Background()

	systemProject := &entities.Project{
		ID:       "system-project-id",
		IsActive: true,
	}

	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		ProjectID: "system-project-id",
		UserID:    nil, // Client JWT
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	targetMember := &entities.Member{
		ID:        "target-member-id",
		UserID:    "target-user-id",
		AccountID: "other-account",
		IsAdmin:   false,
	}

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo) // Once for validation

	// Mock project found
	mockProjectRepo.EXPECT().FindSystem(ctx).Return(systemProject, nil)

	// Mock member found
	mockMemberRepo.EXPECT().FindByID(ctx, "target-member-id").Return(targetMember, nil)

	// Mock admin count (2 admins, so deleting one is OK)
	mockMemberRepo.EXPECT().CountByAccountID(ctx, "other-account", pointy.Bool(true)).Return(2, nil)

	// Mock transaction
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock delete operations - member and user delete succeed, access token delete fails
	mockMemberRepo.EXPECT().Delete(ctx, targetMember).Return(1, nil)
	mockUserRepo.EXPECT().DeleteByID(ctx, "target-user-id").Return(1, nil)
	dbErr := assert.AnError
	mockAccessTokenRepo.EXPECT().DeleteByUserID(ctx, "target-user-id").Return(0, dbErr)

	// Create service
	service := &memberService{
		registry: mockRegistry,
	}

	// Test request
	req := &models.DeleteMemberRequest{
		MemberID: "target-member-id",
	}

	// Execute
	err := service.DeleteMember(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
}
