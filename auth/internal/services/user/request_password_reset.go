package user

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/internal/constants/rediskeys"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/services/shared"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/idutil"
	"github.com/shomali11/util/xhashes"
)

func (s *userService) RequestPasswordReset(ctx context.Context, req *models.RequestPasswordResetRequest) error {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}

	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByIDAndProjectID(ctx, req.UserID, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by ID and project ID", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found by ID and project ID",
			"user_id", req.UserID,
			"project_id", project.ID)
		return errors.ErrUserNotFound
	}
	if user.Email == nil {
		log.WithContext(ctx).Errorw("User has no email address", "user_id", user.ID)
		return errors.ErrUserHasNoEmail
	}
	if !user.IsActive {
		log.WithContext(ctx).Errorw("User is not active", "user_id", user.ID)
		return errors.ErrUserNotActive
	}

	var loginURL *string
	if req.RedirectURI != nil {
		// Validate redirect URI when provided
		redirectURI, err := s.registry.RedirectURIRepository().FindByProjectIDAndRedirectURI(ctx, project.ID, *req.RedirectURI)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find redirect URI by project ID and redirect URI", "error", err, "project_id", project.ID, "redirect_uri", *req.RedirectURI)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if redirectURI == nil {
			log.WithContext(ctx).Errorw("Redirect URI not found", "project_id", project.ID, "redirect_uri", *req.RedirectURI)
			return errors.ErrRedirectURINotFound
		}
		loginURL = redirectURI.LoginURL
	}

	tokenData := models.ResetPasswordTokenData{
		UserID:   user.ID,
		LoginURL: loginURL,
	}

	// Generate reset password token and store in redis
	token := idutil.NanoID(64)
	tokenHash := xhashes.SHA256(token)
	tokenExpiryDuration, err := tparse.AbsoluteDuration(time.Now(), s.defaultConfig.ResetWithContextExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to parse reset password token expiry duration", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ResetPasswordPrefix, tokenHash)
	err = s.redisService.SetData(ctx, redisKey, tokenData, tokenExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to store reset password token in redis", "error", err, "redis_key", redisKey)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	// Publish reset password email job to redis stream
	err = s.publishResetPasswordEmail(ctx, token, user, project)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to publish reset password email", "error", err, "user_id", user.ID, "project_id", project.ID)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	return nil
}
