package member

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/karrick/tparse/v2"
	disposable "github.com/rocketlaunchr/anti-disposable-email"
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

// CreateMember creates a new member for Roled Console (system project). Every member
// created will be a user under the system project associated with the target account.
func (s *memberService) CreateMember(ctx context.Context, req *models.CreateMemberRequest) (*models.CreateMemberResponse, error) {

	// Validate that the access token belongs to the system project
	_, systemProject, err := s.validateMustSystemProject(ctx)
	if err != nil {
		return nil, err
	}

	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
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
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if result == nil {
			log.WithContext(ctx).Errorw("Target account not found", "account_id", targetAccountID)
			return nil, errors.ErrAccountNotFound
		}
		if !result.IsActive {
			log.WithContext(ctx).Errorw("Target account not active", "account_id", targetAccountID)
			return nil, errors.ErrAccountNotActive
		}
	}

	email := strings.ToLower(req.Email)

	// Check if user with the same email already exists under system project
	userRepo := s.registry.UserRepository()
	existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, systemProject.ID, email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by project ID and email", "error", err, "project_id", systemProject.ID, "email", email)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existingUser != nil {
		log.WithContext(ctx).Errorw("User with the same email already exists under the system project",
			"email", email,
			"user_account_id", existingUser.AccountID,
			"req_account_id", req.AccountID)
		return nil, errors.ErrUserEmailAlreadyUsed
	}

	role, err := s.registry.RoleRepository().FindByProjectIDAndCode(ctx, systemProject.ID, constants.RoledConsoleDefaultRoleCode)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find default role by project ID", "error", err, "project_id", systemProject.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if role == nil {
		log.WithContext(ctx).Errorw("Default role not found", "project_id", systemProject.ID)
		return nil, pkgerrors.ErrSystemError.WithError(fmt.Errorf("default role not found"))
	}

	var user *entities.User
	var result *models.CreateMemberResponse
	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Check the system project settings before creating user
		projectSettingRepo := s.registry.ProjectSettingRepository()
		projectSetting, err := projectSettingRepo.FindByProjectID(ctx, systemProject.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find project settings by project ID", "error", err, "project_id", systemProject.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if !projectSetting.IsAllowTempEmail {
			parsedEmail, err := disposable.ParseEmail(req.Email)
			if err != nil || parsedEmail.Disposable {
				log.WithContext(ctx).Errorw("Temporary email addresses are not allowed", "email", req.Email, "error", err)
				return errors.ErrDisposableEmail.WithError(err)
			}
		}

		// Create a user under the system project associated with the target account
		displayName := email[0:strings.Index(email, "@")]
		user = &entities.User{
			ID:          idutil.NewID(),
			AccountID:   targetAccountID,
			ProjectID:   systemProject.ID,
			Email:       &email,
			DisplayName: displayName,
		}
		userRepo := registry.UserRepository()
		err = userRepo.Create(ctx, user)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create user", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Assign default role to the user for system project
		err = registry.UserRoleRepository().Create(ctx, &entities.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		})
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create user role", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create a member record
		member := &entities.Member{
			ID:        idutil.NewID(),
			AccountID: targetAccountID,
			UserID:    user.ID,
		}
		memberRepo := registry.MemberRepository()
		err = memberRepo.Create(ctx, member)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create member", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		result = &models.CreateMemberResponse{
			ID:          member.ID,
			Email:       req.Email, // Return email as provided in the request
			DisplayName: user.DisplayName,
			IsActive:    user.IsActive,
		}
		return nil
	})

	// Send activation email if member is created successfully
	if err == nil {

		var loginURL *string
		if req.RedirectURI != nil {
			// Validate redirect URI when provided
			redirectURI, err := s.registry.RedirectURIRepository().FindByProjectIDAndRedirectURI(ctx, systemProject.ID, *req.RedirectURI)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find redirect URI by project ID and redirect URI", "error", err, "project_id", systemProject.ID, "redirect_uri", *req.RedirectURI)
				return nil, pkgerrors.ErrSystemError.WithError(err)
			}
			if redirectURI == nil {
				log.WithContext(ctx).Errorw("Redirect URI not found", "project_id", systemProject.ID, "redirect_uri", *req.RedirectURI)
				return nil, errors.ErrRedirectURINotFound
			}
			loginURL = redirectURI.LoginURL
		}

		tokenData := models.CreateMemberTokenData{
			UserID:   user.ID,
			LoginURL: loginURL,
		}

		// Generate member activation token
		token := idutil.NanoID(64)
		tokenHash := xhashes.SHA256(token)
		tokenExpiryDuration, err := tparse.AbsoluteDuration(time.Now(), s.defaultConfig.ActivateMemberExpiryDuration)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to parse verify email expiry duration", "duration", s.defaultConfig.ActivateMemberExpiryDuration, "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		// Store token hash in Redis with expiration
		redisKey := fmt.Sprintf("%s:%s", rediskeys.ActivateMemberPrefix, tokenHash)
		err = s.redisService.SetData(ctx, redisKey, tokenData, tokenExpiryDuration)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to store member activation token in redis", "error", err, "token_hash", tokenHash)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}

		// Publish member activation email job to redis stream
		err = s.publishMemberActivationEmail(ctx, token, user, systemProject, account)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to publish member activation email", "error", err, "user_id", user.ID, "project_id", systemProject.ID, "account_id", account.ID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
	}

	return result, err
}

func (s *memberService) publishMemberActivationEmail(ctx context.Context, token string, user *entities.User, project *entities.Project, account *entities.Account) error {
	payload := payloads.EmailPayload{
		Type:            constants.EmailPayloadTypeActivateMember,
		To:              *user.Email,
		From:            fmt.Sprintf(s.defaultConfig.Email.From, project.Name),
		Subject:         fmt.Sprintf("You are invited to join the %s account", account.Name),
		AccountName:     account.Name,
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
