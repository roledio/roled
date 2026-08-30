package account

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"

	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
)

func TestAccountService_DeleteAccount_SystemAccount_SelfDeletion(t *testing.T) {
	ctx := context.Background()
	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "system-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrModifySystemAccount, err)
}

func TestAccountService_DeleteAccount_SystemAccount_OtherAccount_Success(t *testing.T) {
	ctx := context.Background()
	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo) // For top level call
	// AccessTokenRepository is called inside Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		return fn(innerRegistry)
	})

	// Mock account found
	otherAccount := &entities.Account{
		ID:       "other-account",
		IsSystem: false,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "other-account").Return(otherAccount, nil)

	// Mock delete operations
	mockAccountRepo.EXPECT().DeleteByID(ctx, "other-account").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(0, nil)
	mockMemberRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(1, nil)
	mockUserRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(1, nil)

	// Mock cache invalidation
	mockRedis.EXPECT().DeleteManyWithContext(ctx, []string{rediskeys.AccountByID("other-account")}).Return(nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "other-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestAccountService_DeleteAccount_SystemAccount_OtherAccount_SystemTarget(t *testing.T) {
	ctx := context.Background()
	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock account found (system account)
	systemTargetAccount := &entities.Account{
		ID:       "system-target",
		IsSystem: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "system-target").Return(systemTargetAccount, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "system-target",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrModifySystemAccount, err)
}

func TestAccountService_DeleteAccount_SystemAccount_OtherAccount_NotFound(t *testing.T) {
	ctx := context.Background()
	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Mock account not found
	mockAccountRepo.EXPECT().FindByID(ctx, "other-account").Return(nil, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "other-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrAccountNotFound, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_SelfDeletion_UserToken_Success(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo) // For top level call
	// AccessTokenRepository is called inside Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		return fn(innerRegistry)
	})

	// Mock user found with password hash
	passwordHash := "$2a$08$S18QQJOsf3cDrGjHHF0zO.thKaBDU9I45nlmtkzDHOmhmL9G66l7i"
	user := &entities.User{
		ID:           "user-id",
		PasswordHash: &passwordHash,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(user, nil)

	// Mock member found (admin)
	member := &entities.Member{
		ID:        "member-id",
		AccountID: "test-account",
		UserID:    "user-id",
		IsAdmin:   true,
	}
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "user-id").Return(member, nil)

	// Mock delete operations
	mockAccountRepo.EXPECT().DeleteByID(ctx, "test-account").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByAccountID(ctx, "test-account").Return(0, nil)
	mockMemberRepo.EXPECT().DeleteByAccountID(ctx, "test-account").Return(1, nil)
	mockUserRepo.EXPECT().DeleteByAccountID(ctx, "test-account").Return(1, nil)
	mockRedis.EXPECT().DeleteManyWithContext(ctx, []string{rediskeys.AccountByID("test-account")}).Return(nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("correct-password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.NoError(t, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_OtherAccount(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "other-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrAccountNotFound, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_ClientToken(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    nil, // Client token
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNonUserDeleteAccount, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_UserNotFound(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	// Mock user not found
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(nil, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrUserNotFound, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	// Mock user found with password hash
	passwordHash := "$2a$08$S18QQJOsf3cDrGjHHF0zO.thKaBDU9I45nlmtkzDHOmhmL9G66l7i"
	user := &entities.User{
		ID:           "user-id",
		PasswordHash: &passwordHash,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(user, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request with wrong password
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("wrong-password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrInvalidPassword, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_MemberNotFound(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock user found with password hash
	passwordHash := "$2a$08$S18QQJOsf3cDrGjHHF0zO.thKaBDU9I45nlmtkzDHOmhmL9G66l7i"
	user := &entities.User{
		ID:           "user-id",
		PasswordHash: &passwordHash,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(user, nil)

	// Mock member not found
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "user-id").Return(nil, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("correct-password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrMemberNotFound, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_NonAdminUser(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)

	// Mock user found with password hash
	passwordHash := "$2a$08$S18QQJOsf3cDrGjHHF0zO.thKaBDU9I45nlmtkzDHOmhmL9G66l7i"
	user := &entities.User{
		ID:           "user-id",
		PasswordHash: &passwordHash,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(user, nil)

	// Mock member found (not admin)
	member := &entities.Member{
		ID:        "member-id",
		AccountID: "test-account",
		UserID:    "user-id",
		IsAdmin:   false, // Not admin
	}
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "user-id").Return(member, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("correct-password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNonAdminDeleteAccount, err)
}

func TestAccountService_DeleteAccount_AccountNotInContext(t *testing.T) {
	ctx := context.Background()
	// No account in context

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccountNotFound, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_AccessTokenNotInContext(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)
	// No access token in context

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccessTokenNotFound, err)
}

func TestAccountService_DeleteAccount_SystemAccount_DeleteError(t *testing.T) {
	ctx := context.Background()
	systemAccount := &entities.Account{
		ID:       "system-account",
		IsSystem: true,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, systemAccount)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo) // For top level call
	// AccessTokenRepository is called inside Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		innerRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
		innerRegistry.EXPECT().UserRepository().Return(mockUserRepo)
		return fn(innerRegistry)
	})

	// Mock account found
	otherAccount := &entities.Account{
		ID:       "other-account",
		IsSystem: false,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "other-account").Return(otherAccount, nil)

	// Mock delete account error
	mockMemberRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(1, nil)
	mockUserRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(1, nil)
	mockAccessTokenRepo.EXPECT().DeleteByAccountID(ctx, "other-account").Return(0, nil)
	dbErr := assert.AnError
	mockAccountRepo.EXPECT().DeleteByID(ctx, "other-account").Return(0, dbErr)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "other-account",
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
}

func TestAccountService_DeleteAccount_NonSystemAccount_DeleteAccessTokensError(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	accessToken := &entities.AccessToken{
		ID:        "access-token-id",
		UserID:    pointy.String("user-id"),
		AccountID: "test-account",
	}
	ctx = context.WithValue(ctx, constants.CtxAccessToken, accessToken)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockAccessTokenRepo := repositorymocks.NewMockAccessTokenRepository(t)

	// Mock registry to return repositories
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo) // For top level call
	// AccessTokenRepository is called inside Tx
	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		// Create inner registry for transaction
		innerRegistry := repositorymocks.NewMockRegistry(t)
		innerRegistry.EXPECT().AccessTokenRepository().Return(mockAccessTokenRepo)
		return fn(innerRegistry)
	})

	// Mock user found with password hash
	passwordHash := "$2a$08$S18QQJOsf3cDrGjHHF0zO.thKaBDU9I45nlmtkzDHOmhmL9G66l7i"
	user := &entities.User{
		ID:           "user-id",
		PasswordHash: &passwordHash,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-id").Return(user, nil)

	// Mock member found (admin)
	member := &entities.Member{
		ID:        "member-id",
		AccountID: "test-account",
		UserID:    "user-id",
		IsAdmin:   true,
	}
	mockMemberRepo.EXPECT().FindByAccountIDAndUserID(ctx, "test-account", "user-id").Return(member, nil)

	// Mock delete operations
	dbErr := assert.AnError
	mockAccessTokenRepo.EXPECT().DeleteByAccountID(ctx, "test-account").Return(0, dbErr)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.DeleteAccountRequest{
		AccountID: "test-account",
		Password:  pointy.String("correct-password"),
	}

	// Execute
	err := service.DeleteAccount(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
}
