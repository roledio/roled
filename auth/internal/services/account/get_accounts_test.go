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
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"
)

func TestAccountService_GetAccounts_Success_SystemAccount(t *testing.T) {
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

	// Mock count
	mockAccountRepo.EXPECT().Count(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(2, nil)

	// Mock find all
	accounts := []entities.Account{
		{
			ID:          "account1",
			Name:        "Account 1",
			Description: "Desc 1",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "account2",
			Name:        "Account 2",
			Description: "Desc 2",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	mockAccountRepo.EXPECT().FindAll(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(accounts, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, result, 2)
	assert.Equal(t, "account1", result[0].ID)
	assert.Equal(t, "Account 1", result[0].Name)
}

func TestAccountService_GetAccounts_Success_NonSystemAccount(t *testing.T) {
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

	// Mock count with filter
	mockAccountRepo.EXPECT().Count(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), pointy.String("test-account")).Return(1, nil)

	// Mock find all with filter
	accounts := []entities.Account{
		{
			ID:          "test-account",
			Name:        "Test Account",
			Description: "Test Desc",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	mockAccountRepo.EXPECT().FindAll(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), pointy.String("test-account")).Return(accounts, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, result, 1)
	assert.Equal(t, "test-account", result[0].ID)
}

func TestAccountService_GetAccounts_AccountNotInContext(t *testing.T) {
	ctx := context.Background()
	// No account in context

	// Setup mocks
	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockRedis := servicemocks.NewMockRedisService(t)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCtxAccountNotFound, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}

func TestAccountService_GetAccounts_CountError(t *testing.T) {
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

	// Mock count error
	dbErr := assert.AnError
	mockAccountRepo.EXPECT().Count(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(0, dbErr)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}

func TestAccountService_GetAccounts_CountZero(t *testing.T) {
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

	// Mock count zero
	mockAccountRepo.EXPECT().Count(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(0, nil)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Nil(t, result)
}

func TestAccountService_GetAccounts_FindAllError(t *testing.T) {
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

	// Mock count
	mockAccountRepo.EXPECT().Count(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(1, nil)

	// Mock find all error
	dbErr := assert.AnError
	mockAccountRepo.EXPECT().FindAll(ctx, mock.AnythingOfType("*models.GetAccountsRequest"), mock.AnythingOfType("*string")).Return(nil, dbErr)

	// Create service
	service := NewAccountService(mockRegistry, mockRedis)

	// Test request
	req := &models.GetAccountsRequest{}

	// Execute
	result, count, err := service.GetAccounts(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrSystemError.WithError(dbErr), err)
	assert.Nil(t, result)
	assert.Equal(t, 0, count)
}
