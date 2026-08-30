package member

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *memberService) GetMemberDetails(ctx context.Context, req *models.GetMemberDetailsRequest) (*models.GetMemberDetailsResponse, error) {

	// Validate that the access token belongs to the system project
	_, _, err := s.validateMustSystemProject(ctx)
	if err != nil {
		return nil, err
	}

	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
	}

	memberRepo := s.registry.MemberRepository()
	member, err := memberRepo.FindByIDJoinUser(ctx, req.MemberID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find member details", "error", err, "member_id", req.MemberID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if member == nil {
		log.WithContext(ctx).Errorw("Member not found", "member_id", req.MemberID, "account_id", account.ID)
		return nil, errors.ErrMemberNotFound
	}
	if !account.IsSystem && member.AccountID != account.ID {
		log.WithContext(ctx).Errorw("Member does not belong to the current account",
			"member_id", req.MemberID,
			"member_account_id", member.AccountID,
			"account_id", account.ID)
		return nil, errors.ErrMemberNotFound
	}

	response := &models.GetMemberDetailsResponse{
		ID:          member.ID,
		Email:       member.Email,
		DisplayName: member.DisplayName,
		AvatarURL:   member.AvatarURL,
		IsActive:    member.IsActive,
		IsVerified:  member.IsVerified,
		IsAdmin:     member.IsAdmin,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}

	return response, nil
}
