package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/queues/payloads"
	"github.com/roledio/roled/internal/services/shared"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/jsonutil"
)

func (s *userService) ResendVerificationEmail(ctx context.Context, req *models.ResendVerificationEmailRequest) error {
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
	if user.EmailVerifiedAt != nil {
		log.WithContext(ctx).Errorw("User email is already verified", "user_id", user.ID)
		return errors.ErrUserAlreadyVerified
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

	// Publish verification email job to redis stream
	err = s.publishVerificationEmail(ctx, loginURL, user, project)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to publish verification email", "error", err, "user_id", user.ID, "project_id", project.ID)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	return nil
}

func (s *userService) publishVerificationEmail(ctx context.Context, loginURL *string, user *entities.User, project *entities.Project) error {
	payload := payloads.EmailPayload{
		Type:           constants.EmailPayloadTypeVerifyEmail,
		To:             *user.Email,
		From:           fmt.Sprintf(s.defaultConfig.Email.From, project.Name),
		Subject:        "Verify your email for " + project.Name,
		ProjectName:    project.Name,
		ProjectLogoURL: project.LogoURL,
		LoginURL:       loginURL,
		UserID:         user.ID,
		DisplayName:    user.DisplayName,
		IsSignup:       false,
	}
	contextFields := contextutil.GetFields(ctx, constants.RequestLoggerKeys)
	message := queues.Message{
		Payload: jsonutil.Stringify(payload),
		Context: jsonutil.Stringify(contextFields),
	}
	return s.emailPublisher.Publish(ctx, message)
}
