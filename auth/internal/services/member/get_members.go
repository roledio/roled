package member

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *memberService) GetMembers(ctx context.Context, req *models.GetMembersRequest) ([]models.GetMembersResponse, int, error) {

	// Validate that the access token belongs to the system project
	_, _, err := s.validateMustSystemProject(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, 0, errors.ErrCtxAccountNotFound
	}

	// If the account ID is not provided, default to the current account.
	// If it is provided, but the current account is not a system account, override it
	// with the current account ID
	if req.AccountID == "" || !account.IsSystem {
		req.AccountID = account.ID
	}

	targetAccountID := req.AccountID

	// Validate target account if different from current account (only for system accounts)
	if account.ID != targetAccountID {

		accountRepo := s.registry.AccountRepository()
		result, err := accountRepo.FindByID(ctx, targetAccountID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find target account", "error", err, "account_id", targetAccountID)
			return nil, 0, pkgerrors.ErrSystemError.WithError(err)
		}
		if result == nil {
			log.WithContext(ctx).Errorw("Target account not found", "account_id", targetAccountID)
			return nil, 0, errors.ErrAccountNotFound
		}
		if !result.IsActive {
			log.WithContext(ctx).Errorw("Target account not active", "account_id", targetAccountID)
			return nil, 0, errors.ErrAccountNotActive
		}
	}

	memberRepo := s.registry.MemberRepository()
	count, err := memberRepo.Count(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count members", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}
	members, err := memberRepo.FindAll(ctx, req)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find members", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	var resp []models.GetMembersResponse
	for _, member := range members {
		resp = append(resp, models.GetMembersResponse{
			ID:          member.ID,
			Email:       member.Email,
			DisplayName: member.DisplayName,
			AvatarURL:   member.AvatarURL,
			IsActive:    member.IsActive,
			IsVerified:  member.IsVerified,
			IsAdmin:     member.IsAdmin,
			CreatedAt:   member.CreatedAt,
			UpdatedAt:   member.UpdatedAt,
		})
	}
	return resp, count, nil
}
