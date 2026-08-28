package account

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/roledio/roled/internal/models"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAccountService_GetAccountDetails_Success_SelfAccount(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:          "test-account",
		Name:        "Test Account",
		Description: "Test Description",
		IsActive:    true,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountDetailsRequest{
		AccountID: "test-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-account", result.ID)
	assert.Equal(t, "Test Account", result.Name)
}

func TestAccountService_GetAccountDetails_Success_SystemAccountAccessOther(t *testing.T) {
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

	// Mock account found
	otherAccount := &entities.Account{
		ID:          "other-account",
		Name:        "Other Account",
		Description: "Other Description",
		IsActive:    true,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "other-account").Return(otherAccount, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountDetailsRequest{
		AccountID: "other-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "other-account", result.ID)
	assert.Equal(t, "Other Account", result.Name)
}

func TestAccountService_GetAccountDetails_AccountNotInContext(t *testing.T) {
	ctx := context.Background()
	// No account in context

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountDetailsRequest{
		AccountID: "test-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccountNotFound, err)
	assert.Nil(t, result)
}

func TestAccountService_GetAccountDetails_NonSystemAccessOtherAccount(t *testing.T) {
	ctx := context.Background()
	account := &entities.Account{
		ID:       "test-account",
		IsSystem: false,
	}
	ctx = context.WithValue(ctx, constants.CtxAccount, account)

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountDetailsRequest{
		AccountID: "other-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrAccountNotFound, err)
	assert.Nil(t, result)
}

func TestAccountService_GetAccountDetails_SystemAccountOtherNotFound(t *testing.T) {
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
	req := &models.GetAccountDetailsRequest{
		AccountID: "other-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrAccountNotFound, err)
	assert.Nil(t, result)
}

func TestAccountService_GetAccountDetails_SystemAccountOtherRepoError(t *testing.T) {
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

	// Mock account repo error
	dbErr := assert.AnError
	mockAccountRepo.EXPECT().FindByID(ctx, "other-account").Return(nil, dbErr)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountDetailsRequest{
		AccountID: "other-account",
	}

	// Execute
	result, err := service.GetAccountDetails(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
	assert.Nil(t, result)
}
