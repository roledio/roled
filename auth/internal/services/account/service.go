package account

import (
	"context"

	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
)

type AccountService interface {
	GetAccountDetails(ctx context.Context, req *models.GetAccountDetailsRequest) (*models.GetAccountDetailsResponse, error)
	GetAccounts(ctx context.Context, req *models.GetAccountsRequest) ([]models.GetAccountsResponse, int, error)
	UpdateAccount(ctx context.Context, req *models.UpdateAccountRequest) (*models.UpdateAccountResponse, error)
	DeleteAccount(ctx context.Context, req *models.DeleteAccountRequest) error
}

type accountService struct {
	registry repositories.Registry
	redis    infra.RedisService
}

func NewAccountService(registry repositories.Registry, redis infra.RedisService) AccountService {
	return &accountService{
		registry: registry,
		redis:    redis,
	}
}
