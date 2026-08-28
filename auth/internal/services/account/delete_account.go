package account

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/shared"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/passwordutil"
	"go.openly.dev/pointy"
)

func (s *accountService) DeleteAccount(ctx context.Context, req *models.DeleteAccountRequest) error {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return errors.ErrCtxAccountNotFound
	}

	accountRepo := s.registry.AccountRepository()

	if account.IsSystem {

		// The account id in the request is the same as the system account id
		// Prevent deletion of system account
		if account.ID == req.AccountID {
			log.WithContext(ctx).Errorw("System account deletion not allowed", "account_id", account.ID)
			return errors.ErrModifySystemAccount
		}

		// System account can delete other accounts, find the target account
		result, err := accountRepo.FindByID(ctx, req.AccountID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find account by id", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if result == nil {
			log.WithContext(ctx).Errorw("Account not found", "req_account_id", req.AccountID)
			return errors.ErrAccountNotFound
		}
		if result.IsSystem { // Won't happen since there is only one system account
			log.WithContext(ctx).Errorw("Cannot delete system account", "req_account_id", req.AccountID)
			return errors.ErrModifySystemAccount
		}
		account = result

	} else {

		err := s.validateNonSystemAccountDeletion(ctx, account, req)
		if err != nil {
			return err
		}
	}

	err := s.registry.Tx(func(registry repositories.Registry) error {

		// Delete all access tokens associated with this account
		accessTokenRepo := registry.AccessTokenRepository()
		_, err := accessTokenRepo.DeleteByAccountID(ctx, account.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete access tokens by account ID", "error", err, "account_id", account.ID)
			return err
		}

		// Delete all members associated with this account
		memberRepo := registry.MemberRepository()
		_, err = memberRepo.DeleteByAccountID(ctx, account.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete members by account ID", "error", err, "account_id", account.ID)
			return err
		}

		// Delete all users associated with this account
		userRepo := registry.UserRepository()
		_, err = userRepo.DeleteByAccountID(ctx, account.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete users by account ID", "error", err, "account_id", account.ID)
			return err
		}

		// Delete the account
		accountRepo := registry.AccountRepository()
		affected, err := accountRepo.DeleteByID(ctx, account.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete account", "error", err, "account_id", account.ID)
			return err
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No account deleted", "account_id", account.ID)
			return errors.ErrAccountNotFound
		}

		return nil
	})

	if err == nil {
		// Invalidate cache after successful deletion
		shared.InvalidateAccountCache(ctx, s.redis, account.ID)
	}

	return err
}

func (s *accountService) validateNonSystemAccountDeletion(ctx context.Context, account *entities.Account, req *models.DeleteAccountRequest) error {

	// Non system account can only delete their own account
	// When account IDs don't match, return not found to prevent information leakage
	if account.ID != req.AccountID {
		log.WithContext(ctx).Errorw("Non system account trying to delete other account", "current_account_id", account.ID, "req_account_id", req.AccountID)
		return errors.ErrAccountNotFound
	}

	// Get current access token from context, should not be nil here
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		return errors.ErrCtxAccessTokenNotFound
	}

	// Non system account can delete their account only with user JWT (from authorization code flow)
	// and must provide valid password
	if accessToken.UserID == nil {
		log.WithContext(ctx).Errorw("Non user token attempt to delete account", "access_token_id", accessToken.ID)
		return errors.ErrNonUserDeleteAccount
	}

	userID := *accessToken.UserID
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID", "error", err, "user_id", userID)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", userID)
		return errors.ErrUserNotFound
	}
	// Users of system project must have a password
	if user.PasswordHash == nil {
		log.WithContext(ctx).Errorw("User has no password hash which is unexpected", "user_id", userID)
		return pkgerrors.ErrSystemError
	}
	password := pointy.StringValue(req.Password, "")
	if !passwordutil.IsValidPassword(password, *user.PasswordHash) {
		log.WithContext(ctx).Errorw("Invalid password for user", "user_id", userID)
		return errors.ErrInvalidPassword
	}

	// Validate if the user is an admin of the account
	memberRepo := s.registry.MemberRepository()
	member, err := memberRepo.FindByAccountIDAndUserID(ctx, account.ID, userID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find member by account ID and user ID", "error", err, "account_id", account.ID, "user_id", userID)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if member == nil {
		log.WithContext(ctx).Errorw("Member not found by account ID and user ID", "account_id", account.ID, "user_id", userID)
		return errors.ErrMemberNotFound
	}
	if !member.IsAdmin {
		log.WithContext(ctx).Errorw("User is not an admin of the account", "account_id", account.ID, "user_id", userID)
		return errors.ErrNonAdminDeleteAccount
	}
	return nil
}
