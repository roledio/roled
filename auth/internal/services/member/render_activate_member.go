package member

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/shomali11/util/xhashes"
)

func (s *memberService) RenderActivateMember(ctx context.Context, req *models.RenderActivateMemberRequest) (*models.RenderActivateMemberResponse, error) {
	if req.UserID != nil {
		// If the user ID is provided, skip token validation
		// This is used when user has successfully activated their account (token is already deleted from redis)
		// and is redirected to the activate page again
		res, err := s.prepareRenderActivateMember(ctx, *req.UserID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	tokenHash := xhashes.SHA256(req.Token)
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateMemberPrefix, tokenHash)
	tokenData, err := s.readTokenData(ctx, redisKey)
	if err != nil {
		return nil, err
	}

	res, err := s.prepareRenderActivateMember(ctx, tokenData.UserID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *memberService) readTokenData(ctx context.Context, redisKey string) (*models.CreateMemberTokenData, error) {
	tokenData := models.CreateMemberTokenData{}
	found, err := s.redisService.GetData(ctx, redisKey, &tokenData)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to get activate member token from redis", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if !found {
		log.WithContext(ctx).Error("Activate member token not found")
		return nil, errors.ErrInvalidActivateMemberToken
	}
	log.WithContext(ctx).Debugw("Activate member token found", "key", redisKey, "user_id", tokenData.UserID)
	return &tokenData, nil
}

func (s *memberService) prepareRenderActivateMember(ctx context.Context, userID string) (*models.RenderActivateMemberResponse, error) {
	// Validate user
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID", "error", err, "user_id", userID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "user_id", userID)
		return nil, errors.ErrUserNotFound
	}

	// Validate account
	accountRepo := s.registry.AccountRepository()
	account, err := accountRepo.FindByID(ctx, user.AccountID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find account by ID", "error", err, "account_id", user.AccountID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if account == nil {
		log.WithContext(ctx).Errorw("Account not found", "account_id", user.AccountID)
		return nil, errors.ErrAccountNotFound
	}
	if !account.IsActive {
		log.WithContext(ctx).Errorw("Account is not active", "account_id", user.AccountID)
		return nil, errors.ErrAccountNotActive
	}

	// Validate project
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, user.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err, "project_id", user.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", user.ProjectID)
		return nil, errors.ErrProjectNotFound
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "project_id", user.ProjectID)
		return nil, errors.ErrProjectNotActive
	}

	return &models.RenderActivateMemberResponse{
		User:    user,
		Account: account,
		Project: project,
	}, nil
}
