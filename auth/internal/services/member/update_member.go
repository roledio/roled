package member

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

// UpdateMember updates the is_admin status of a member.
//
// Only the is_admin field is updated; other member fields are not affected.
// If req.IsAdmin is nil, the request is a no-op and the current state is returned.
func (s *memberService) UpdateMember(ctx context.Context, req *models.UpdateMemberRequest) (*models.UpdateMemberResponse, error) {

	// Validate that the access token belongs to the system project
	accessToken, _, err := s.validateMustSystemProject(ctx)
	if err != nil {
		return nil, err
	}

	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
	}

	// Route to appropriate handler based on JWT type
	var member *entities.Member
	if accessToken.UserID == nil {
		// Client JWT
		member, err = s.handleUpdateMemberUsingClientJWT(ctx, req, account)
	} else {
		// User JWT
		member, err = s.handleUpdateMemberUsingUserJWT(ctx, req, account, accessToken)
	}
	if err != nil {
		return nil, err
	}

	// If is_admin is nil, nothing changes — return current state as-is
	if req.IsAdmin == nil {
		return &models.UpdateMemberResponse{
			ID:        member.ID,
			CreatedAt: member.CreatedAt,
			UpdatedAt: member.UpdatedAt,
			AccountID: member.AccountID,
			UserID:    member.UserID,
			IsAdmin:   member.IsAdmin,
		}, nil
	}

	// Persist the is_admin change
	member.IsAdmin = *req.IsAdmin
	member.UpdatedAt = time.Now().UTC()
	memberRepo := s.registry.MemberRepository()
	affected, err := memberRepo.Update(ctx, member)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to update member", "error", err, "member_id", member.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if affected == 0 {
		log.WithContext(ctx).Errorw("No rows affected when updating member", "member_id", member.ID)
		return nil, errors.ErrMemberNotFound
	}

	// Invalidate cache after successful update
	shared.InvalidateMemberCache(ctx, s.redisService, member)

	return &models.UpdateMemberResponse{
		ID:        member.ID,
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
		AccountID: member.AccountID,
		UserID:    member.UserID,
		IsAdmin:   member.IsAdmin,
	}, nil
}

// handleUpdateMemberUsingClientJWT handles update member request when the access token is a client JWT.
//
//  1. System account is allowed to update any member.
//     1.1. If account_id is provided, validate that the target member belongs to that account.
//  2. Non system account:
//     2.1. Check if the target member belongs to the same account.
//     2.2. If demoting (is_admin = false) and target member is currently admin, check at least 2 admins exist.
func (s *memberService) handleUpdateMemberUsingClientJWT(ctx context.Context, req *models.UpdateMemberRequest, account *entities.Account) (*entities.Member, error) {
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

	// System account: if account_id is provided, validate that the target member belongs to that account
	if account.IsSystem && req.AccountID != "" && targetMember.AccountID != req.AccountID {
		log.WithContext(ctx).Errorw("Target member does not belong to specified account",
			"target_account_id", targetMember.AccountID,
			"req_account_id", req.AccountID)
		return nil, errors.ErrMemberNotFound
	}

	// Non system account can only update members from the same account
	if !account.IsSystem && targetMember.AccountID != account.ID {
		log.WithContext(ctx).Errorw("Non system account trying to update other account member",
			"target_account_id", targetMember.AccountID,
			"req_account_id", account.ID)
		return nil, errors.ErrMemberNotFound
	}

	// Guard against demoting the last admin
	if err := s.guardDemoteLastAdmin(ctx, memberRepo, targetMember, req.IsAdmin); err != nil {
		return nil, err
	}

	return targetMember, nil
}

// handleUpdateMemberUsingUserJWT handles update member request when the access token is a user JWT.
//
//  1. System account is allowed to update any member.
//     1.1. Cannot update self.
//     1.2. Guard against demoting the last admin.
//  2. Non system account:
//     2.1. Check if the target member belongs to the same account.
//     2.2. Caller must be an admin member.
//     2.3. Cannot update self.
//     2.4. Guard against demoting the last admin.
func (s *memberService) handleUpdateMemberUsingUserJWT(ctx context.Context, req *models.UpdateMemberRequest, account *entities.Account,
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

		// Non system account can only update members from the same account
		if targetMember.AccountID != account.ID {
			log.WithContext(ctx).Errorw("Non system account trying to update other account member",
				"target_account_id", targetMember.AccountID,
				"req_account_id", account.ID)
			return nil, errors.ErrMemberNotFound
		}

		// Non system account member must be admin to update other members
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
			log.WithContext(ctx).Errorw("Non admin member trying to update other members",
				"account_id", accessToken.AccountID,
				"member_id", currentMember.ID)
			return nil, errors.ErrNonAdminUpdateMember
		}
	}

	currentUserID := *accessToken.UserID

	// Cannot update self (admin status self-change is not allowed)
	if targetMember.UserID == currentUserID {
		log.WithContext(ctx).Errorw("Cannot update own admin status",
			"member_id", targetMember.ID,
			"user_id", currentUserID)
		return nil, errors.ErrCannotUpdateSelf
	}

	// Guard against demoting the last admin
	if err := s.guardDemoteLastAdmin(ctx, memberRepo, targetMember, req.IsAdmin); err != nil {
		return nil, err
	}

	return targetMember, nil
}

// guardDemoteLastAdmin returns an error if the caller is attempting to demote the last admin of an account.
func (s *memberService) guardDemoteLastAdmin(ctx context.Context, memberRepo interface {
	CountByAccountID(ctx context.Context, accountID string, isAdmin *bool) (int, error)
}, targetMember *entities.Member, isAdmin *bool) error {
	// Only relevant when demoting (setting is_admin to false) an existing admin
	if isAdmin == nil || *isAdmin || !targetMember.IsAdmin {
		return nil
	}
	isAdminFilter := true
	adminCount, err := memberRepo.CountByAccountID(ctx, targetMember.AccountID, &isAdminFilter)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count admin members by account id", "error", err, "account_id", targetMember.AccountID)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if adminCount <= 1 {
		log.WithContext(ctx).Errorw("Cannot demote the last admin member",
			"account_id", targetMember.AccountID,
			"member_id", targetMember.ID)
		return errors.ErrCannotDemoteLastAdmin
	}
	return nil
}
