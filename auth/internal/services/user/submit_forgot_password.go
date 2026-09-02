package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues"
	"github.com/roledio/roled/auth/internal/queues/payloads"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/roledio/roled/auth/pkg/utils/jsonutil"
	"github.com/shomali11/util/xhashes"
)

func (s *userService) SubmitForgotPassword(ctx context.Context, req *models.SubmitForgotPasswordRequest) error {
	project, loginURL, err := s.validateForgotPassword(ctx, req.ClientID, req.RedirectURI)
	if err != nil {
		return err
	}
	email := strings.ToLower(req.Email) // Lowercase email before lookup
	userRepo := s.registry.UserRepository()
	user, err := userRepo.FindByProjectIDAndEmail(ctx, project.ID, email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by email and project ID", "error", err, "email", req.Email, "project_id", project.ID)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		// To prevent email enumeration, we will not return error if user not found
		log.WithContext(ctx).Warnw("User not found with given email and project ID, no email will be sent", "email", req.Email, "project_id", project.ID)
		return nil
	}
	if !user.IsActive {
		log.WithContext(ctx).Warnw("User is not active, no email will be sent", "user_id", user.ID)
		return nil
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

func (s *userService) publishResetPasswordEmail(ctx context.Context, token string, user *entities.User, project *entities.Project) error {
	payload := payloads.EmailPayload{
		Type:            constants.EmailPayloadTypeResetPassword,
		To:              *user.Email,
		From:            fmt.Sprintf(s.defaultConfig.Email.From, project.Name),
		Subject:         "Reset your password for " + project.Name,
		ProjectName:     project.Name,
		ProjectLogoURL:  project.LogoURL,
		ProjectIsSystem: project.IsSystem,
		DisplayName:     user.DisplayName,
		Token:           token,
	}
	contextFields := contextutil.GetFields(ctx, constants.RequestLoggerKeys)
	message := queues.Message{
		Payload: jsonutil.Stringify(payload),
		Context: jsonutil.Stringify(contextFields),
	}
	return s.emailPublisher.Publish(ctx, message)
}
