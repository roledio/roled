package account

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *accountService) GetAccounts(ctx context.Context, req *models.GetAccountsRequest) ([]models.GetAccountsResponse, int, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, 0, errors.ErrCtxAccountNotFound
	}

	// System account can see all accounts, non system account can only see their own accounts
	var filterAccountID *string
	if !account.IsSystem {
		filterAccountID = &account.ID
	}
	accountRepo := s.registry.AccountRepository()
	count, err := accountRepo.Count(ctx, req, filterAccountID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count accounts", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}
	accounts, err := accountRepo.FindAll(ctx, req, filterAccountID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find accounts", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	var resp []models.GetAccountsResponse
	for _, account := range accounts {
		resp = append(resp, models.GetAccountsResponse{
			ID:          account.ID,
			Name:        account.Name,
			Description: account.Description,
			IsActive:    account.IsActive,
			CreatedAt:   account.CreatedAt,
			UpdatedAt:   account.UpdatedAt,
		})
	}
	return resp, count, nil
}
