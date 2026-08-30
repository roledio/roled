package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
)

type AccountRepository interface {
	FindByID(ctx context.Context, id string) (*entities.Account, error)
	FindAll(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) ([]entities.Account, error)
	Count(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) (int, error)
	Create(ctx context.Context, account *entities.Account) error
	Update(ctx context.Context, account *entities.Account) (int, error)
	DeleteByID(ctx context.Context, id string) (int, error)
}
