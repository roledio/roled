package member

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
)

func (s *memberService) DeleteMember(ctx context.Context, req *models.DeleteMemberRequest) error {

	// Validate that the access token belongs to the system project
	accessToken, _, err := s.validateMustSystemProject(ctx)
	if err != nil {
		return err
	}

	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return errors.ErrCtxAccountNotFound
	}

	// Route to appropriate handler based on JWT type
	var member *entities.Member
	if accessToken.UserID == nil {
		// Client JWT
		member, err = s.handleDeleteMemberUsingClientJWT(ctx, req, account)
	} else {
		// User JWT
		member, err = s.handleDeleteMemberUsingUserJWT(ctx, req, account, accessToken)
	}
	if err != nil {
		return err
	}

	// Proceed to delete the member, user, and associated access tokens in a transaction
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Delete member record
		memberRepo := registry.MemberRepository()
		affected, err := memberRepo.Delete(ctx, member)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete member", "error", err, "member_id", member.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No rows affected when deleting member", "member_id", member.ID)
			return errors.ErrMemberNotFound
		}

		// Delete user record
		userRepo := registry.UserRepository()
		_, err = userRepo.DeleteByID(ctx, member.UserID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete user", "error", err, "user_id", member.UserID)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Delete access tokens associated with the user
		accessTokenRepo := registry.AccessTokenRepository()
		_, err = accessTokenRepo.DeleteByUserID(ctx, member.UserID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete access tokens by user ID", "error", err, "user_id", member.UserID)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		return nil
	})

	if err == nil {
		// Invalidate member cache after successful deletion
		shared.InvalidateMemberCache(ctx, s.redisService, member)
	}

	return err
}

// handleDeleteMemberUsingClientJWT handles delete member request when the access token is a client JWT.
//
//  1. System account is allowed to delete any member.
//     1.1. Calculate admin members count, if only one admin member left, prevent deletion
//     1.2. Proceed to delete the member
//  2. Non system account:
//     2.1. Check if the target member belongs to the same account
//     2.2. Calculate admin members count, if only one admin member left, prevent deletion
//     2.3. Proceed to delete the member
func (s *memberService) handleDeleteMemberUsingClientJWT(ctx context.Context, req *models.DeleteMemberRequest, account *entities.Account) (*entities.Member, error) {
	// Find the target member
	memberRepo := s.registry.MemberRepository()
	targetMember, err := memberRepo.FindByID(ctx, req.MemberID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find member by ID", "error", err, "member_id", req.MemberID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if targetMember == nil {
		log.WithContext(ctx).Errorw("Target member not found", "member_id", req.MemberID)
		return nil, errors.ErrMemberNotFound
	}

	// Non system account can only delete members from the same account
	if !account.IsSystem && targetMember.AccountID != account.ID {
		log.WithContext(ctx).Errorw("Non system account trying to delete other account member",
			"target_account_id", targetMember.AccountID,
			"req_account_id", account.ID)
		return nil, errors.ErrMemberNotFound
	}

	// Check if the target member is the last admin
	isAdmin := true
	adminCount, err := memberRepo.CountByAccountID(ctx, targetMember.AccountID, &isAdmin)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count member by account id", "error", err, "account_id", targetMember.AccountID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if targetMember.IsAdmin && adminCount <= 1 {
		log.WithContext(ctx).Errorw("Cannot delete the last admin member",
			"account_id", targetMember.AccountID,
			"member_id", targetMember.ID)
		return nil, errors.ErrCannotDeleteLastAdmin
	}
	return targetMember, nil
}

// handleDeleteMemberUsingUserJWT handles delete member request when the access token is a user JWT.
//
//  1. System account is allowed to delete any member.
//     1.1. If the target member user id is the same as the access token user id, prevent deletion
//     1.2. Calculate admin members count, if only one admin member left, prevent deletion
//     1.3. Proceed to delete the member
//  2. Non system account:
//     2.1. Check if the target member belongs to the same account
//     2.2. Member is not admin, prevent deletion
//     2.3. If the target member user id is the same as the access token user id, prevent deletion
//     2.4. Calculate admin members count, if only one admin member left, prevent deletion
//     2.5. Proceed to delete the member
func (s *memberService) handleDeleteMemberUsingUserJWT(ctx context.Context, req *models.DeleteMemberRequest, account *entities.Account,
	accessToken *entities.AccessToken) (*entities.Member, error) {

	// Find the target member
	memberRepo := s.registry.MemberRepository()
	targetMember, err := memberRepo.FindByID(ctx, req.MemberID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find member by ID", "error", err, "member_id", req.MemberID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if targetMember == nil {
		log.WithContext(ctx).Errorw("Target member not found", "member_id", req.MemberID)
		return nil, errors.ErrMemberNotFound
	}

	if !account.IsSystem {

		// Non system account can only delete members from the same account
		if targetMember.AccountID != account.ID {
			log.WithContext(ctx).Errorw("Non system account trying to delete other account member",
				"target_account_id", targetMember.AccountID,
				"req_account_id", account.ID)
			return nil, errors.ErrMemberNotFound
		}

		// Non system account member must be admin to delete other members
		currentMember, err := memberRepo.FindByAccountIDAndUserID(ctx, accessToken.AccountID, *accessToken.UserID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find current member by account ID and user ID", "error", err, "account_id", accessToken.AccountID, "user_id", *accessToken.UserID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if currentMember == nil {
			log.WithContext(ctx).Errorw("Current member not found for account ID and user ID", "account_id", accessToken.AccountID, "user_id", *accessToken.UserID)
			return nil, errors.ErrMemberNotFound
		}
		if !currentMember.IsAdmin {
			log.WithContext(ctx).Errorw("Non admin member trying to delete other members",
				"account_id", accessToken.AccountID,
				"member_id", currentMember.ID)
			return nil, errors.ErrNonAdminDeleteMember
		}
	}

	currentUserID := *accessToken.UserID

	// Cannot delete self
	if targetMember.UserID == currentUserID {
		log.WithContext(ctx).Errorw("Cannot delete self member",
			"member_id", targetMember.ID,
			"user_id", currentUserID)
		return nil, errors.ErrCannotDeleteSelf
	}

	// Check if the target member is the last admin
	isAdmin := true
	adminCount, err := memberRepo.CountByAccountID(ctx, targetMember.AccountID, &isAdmin)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count member by account id", "error", err, "account_id", targetMember.AccountID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if targetMember.IsAdmin && adminCount <= 1 {
		log.WithContext(ctx).Errorw("Cannot delete the last admin member",
			"account_id", targetMember.AccountID,
			"member_id", targetMember.ID)
		return nil, errors.ErrCannotDeleteLastAdmin
	}

	return targetMember, nil
}
