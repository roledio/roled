package account

import (
	"context"

	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgredis "github.com/roledio/roled/auth/pkg/redis"
)

type AccountService interface {
	GetAccountDetails(ctx context.Context, req *models.GetAccountDetailsRequest) (*models.GetAccountDetailsResponse, error)
	GetAccounts(ctx context.Context, req *models.GetAccountsRequest) ([]models.GetAccountsResponse, int, error)
	UpdateAccount(ctx context.Context, req *models.UpdateAccountRequest) (*models.UpdateAccountResponse, error)
	DeleteAccount(ctx context.Context, req *models.DeleteAccountRequest) error
}

type accountService struct {
	registry repositories.Registry
	redis    pkgredis.Service
}

func NewAccountService(registry repositories.Registry, redis pkgredis.Service) AccountService {
	return &accountService{
		registry: registry,
		redis:    redis,
	}
}
