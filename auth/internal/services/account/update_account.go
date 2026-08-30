package account

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *accountService) UpdateAccount(ctx context.Context, req *models.UpdateAccountRequest) (*models.UpdateAccountResponse, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
	}

	accountRepo := s.registry.AccountRepository()

	if account.IsSystem {

		// The account id in the request is the same as the system account id
		// Prevent modification of system account
		if account.ID == req.AccountID {
			log.WithContext(ctx).Errorw("System account modification not allowed", "account_id", account.ID)
			return nil, errors.ErrModifySystemAccount
		}

		// System account can modify other accounts, find the target account
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
		account.IsActive = req.IsActive // Active status can only be modified by system account

	} else if account.ID != req.AccountID {
		// Non system account can only modify their own account
		// When account IDs don't match, return not found to prevent information leakage
		log.WithContext(ctx).Errorw("Non system account trying to modify other account", "account_id", account.ID, "req_account_id", req.AccountID)
		return nil, errors.ErrAccountNotFound
	}

	account.Name = req.Name
	account.Description = req.Description
	account.UpdatedAt = time.Now().UTC()
	affected, err := accountRepo.Update(ctx, account)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update account", "error", err)
		return nil, err
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No account updated", "account_id", account.ID)
		return nil, errors.ErrAccountNotFound
	}

	// Invalidate cache after successful update
	shared.InvalidateAccountCache(ctx, s.redis, account.ID)

	resp := models.UpdateAccountResponse{
		ID:          account.ID,
		Name:        account.Name,
		Description: account.Description,
		IsActive:    account.IsActive,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}

	return &resp, nil
}
