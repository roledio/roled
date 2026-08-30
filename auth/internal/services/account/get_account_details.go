package account

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *accountService) GetAccountDetails(ctx context.Context, req *models.GetAccountDetailsRequest) (*models.GetAccountDetailsResponse, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
	}

	if account.ID != req.AccountID {

		// Non system account trying to access other account, return not found
		// to prevent information leakage
		if !account.IsSystem {
			log.WithContext(ctx).Errorw("Non system account trying to access other account", "account_id", account.ID, "req_account_id", req.AccountID)
			return nil, errors.ErrAccountNotFound
		}

		accountRepo := s.registry.AccountRepository()
		result, err := accountRepo.FindByID(ctx, req.AccountID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find account by id", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if result == nil {
			log.WithContext(ctx).Errorw("Account not found", "req_account_id", req.AccountID)
			return nil, errors.ErrAccountNotFound
		}
		account = result
	}
	resp := &models.GetAccountDetailsResponse{
		ID:          account.ID,
		Name:        account.Name,
		Description: account.Description,
		IsActive:    account.IsActive,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
	return resp, nil
}
