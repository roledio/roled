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
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/queues"
	"github.com/roledio/roled/auth/internal/queues/payloads"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/roledio/roled/auth/pkg/utils/jsonutil"
	"github.com/shomali11/util/xhashes"
)

// InviteUser invites a user to a project. Creates a new user record with is_active=false
// and sends an email invitation with activation token.
func (s *userService) InviteUser(ctx context.Context, req *models.InviteUserRequest) (*models.UserDetails, error) {

	// Get current access token and project from context
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		log.WithContext(ctx).Errorw("Access token not found in context")
		return nil, errors.ErrCtxAccessTokenNotFound
	}

	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project", "error", err, "project_id", req.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "project_id", req.ProjectID)
		return nil, errors.ErrProjectNotFound
	}

	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project not active", "project_id", req.ProjectID)
		return nil, errors.ErrProjectNotActive
	}

	email := strings.ToLower(req.Email)

	// Check if user with the same email already exists in the project
	userRepo := s.registry.UserRepository()
	existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, req.ProjectID, email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by project ID and email", "error", err, "project_id", req.ProjectID, "email", email)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existingUser != nil {
		log.WithContext(ctx).Errorw("User with the same email already exists in the project", "email", email, "project_id", req.ProjectID)
		return nil, errors.ErrUserEmailAlreadyUsed
	}

	// Validate role if provided
	var role *entities.Role
	if req.RoleID != nil {
		roleRepo := s.registry.RoleRepository()
		r, err := roleRepo.FindByIDAndProjectID(ctx, *req.RoleID, req.ProjectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find role", "error", err, "role_id", *req.RoleID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if r == nil {
			log.WithContext(ctx).Errorw("Role not found", "role_id", *req.RoleID)
			return nil, errors.ErrRoleNotFound
		}
		role = r
	}

	// Validate redirect URI if provided
	var loginURL *string
	if req.RedirectURI != nil {
		redirectURIRepo := s.registry.RedirectURIRepository()
		redirectURI, err := redirectURIRepo.FindByProjectIDAndRedirectURI(ctx, req.ProjectID, *req.RedirectURI)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find redirect URI by project ID and redirect URI", "error", err, "project_id", req.ProjectID, "redirect_uri", *req.RedirectURI)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if redirectURI == nil {
			log.WithContext(ctx).Errorw("Redirect URI not found", "project_id", req.ProjectID, "redirect_uri", *req.RedirectURI)
			return nil, errors.ErrRedirectURINotFound
		}
		loginURL = redirectURI.LoginURL
	}

	// Get current account from context
	account := contextutil.GetAccount(ctx)
	if account == nil {
		log.WithContext(ctx).Errorw("Account not found in context")
		return nil, errors.ErrCtxAccountNotFound
	}

	// Create user and user role in transaction
	var user *entities.User
	displayName := email[0:strings.Index(email, "@")]
	err = s.registry.Tx(func(registry repositories.Registry) error {
		// Create user entity
		user = &entities.User{
			ID:          idutil.NewID(),
			AccountID:   account.ID,
			ProjectID:   req.ProjectID,
			Email:       &email,
			DisplayName: displayName,
			IsActive:    false, // User is inactive until they activate their account
			// EmailVerifiedAt is nil by default
		}

		userRepo := registry.UserRepository()
		if err := userRepo.Create(ctx, user); err != nil {
			log.WithContext(ctx).Errorw("Failed to create user", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Assign role to the user if provided
		if role != nil {
			if err := registry.UserRoleRepository().Create(ctx, &entities.UserRole{
				UserID: user.ID,
				RoleID: role.ID,
			}); err != nil {
				log.WithContext(ctx).Errorw("Failed to create user role", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send invitation email
	err = s.publishInviteUserEmail(ctx, user, project, account, loginURL)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to publish invite user email", "error", err, "user_id", user.ID, "project_id", req.ProjectID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	response := &models.UserDetails{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}

	return response, nil
}

func (s *userService) publishInviteUserEmail(ctx context.Context, user *entities.User, project *entities.Project, account *entities.Account, loginURL *string) error {
	// Generate activation token
	token := idutil.NanoID(64)
	tokenHash := xhashes.SHA256(token)
	tokenExpiryDuration, err := tparse.AbsoluteDuration(time.Now(), s.defaultConfig.ActivateMemberExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to parse activate user expiry duration", "duration", s.defaultConfig.ActivateMemberExpiryDuration, "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	// Store token data in Redis with expiration
	tokenData := models.ActivateProjectUserTokenData{
		UserID:   user.ID,
		LoginURL: loginURL,
	}
	redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateProjectUserPrefix, tokenHash)
	err = s.redisService.SetData(ctx, redisKey, tokenData, tokenExpiryDuration)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to store project user activation token in redis", "error", err, "token_hash", tokenHash)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	// Publish invitation email job to redis stream
	payload := payloads.EmailPayload{
		Type:            constants.EmailPayloadTypeInviteUser,
		To:              *user.Email,
		From:            fmt.Sprintf(s.defaultConfig.Email.From, project.Name),
		Subject:         fmt.Sprintf("You are invited to join %s", project.Name),
		AccountName:     account.Name,
		ProjectName:     project.Name,
		ProjectLogoURL:  project.LogoURL,
		ProjectIsSystem: project.IsSystem,
		DisplayName:     user.DisplayName,
		UserID:          user.ID,
		LoginURL:        loginURL,
		Token:           token,
	}

	contextFields := contextutil.GetFields(ctx, constants.RequestLoggerKeys)
	message := queues.Message{
		Payload: jsonutil.Stringify(payload),
		Context: jsonutil.Stringify(contextFields),
	}
	return s.emailPublisher.Publish(ctx, message)
}
