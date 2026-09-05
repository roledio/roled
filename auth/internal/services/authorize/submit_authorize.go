package authorize

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	disposable "github.com/rocketlaunchr/anti-disposable-email"
	"github.com/roledio/roled/auth/internal/constants"
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
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
)

func (s *authorizeService) SubmitAuthorize(ctx context.Context, req *models.SubmitAuthorizeRequest) (*models.SubmitAuthorizeResult, error) {
	project, redirectURI, projectSetting, err := s.validateAuthorizeRequest(ctx, &req.RenderAuthorizeRequest)
	if err != nil {
		return nil, err
	}

	var user *entities.User
	if req.IsSignup {
		err := s.validateSignup(ctx, project, projectSetting, req)
		if err != nil {
			return nil, err
		}
	} else {
		user, err = s.validateSignin(ctx, project, req)
		if err != nil {
			return nil, err
		}
	}

	var code string
	err = s.registry.Tx(func(registry repositories.Registry) error {
		if req.IsSignup {
			// Process signup: create account, user, and assign default role
			user, err = s.processSignup(ctx, registry, project, projectSetting, redirectURI, req)
			if err != nil {
				return err
			}
		}

		generated, authCode := s.buildAuthCode(ctx, user, project, &req.RenderAuthorizeRequest)
		authCodeRepo := registry.AuthCodeRepository()
		err = authCodeRepo.Create(ctx, authCode)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create auth code", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		code = generated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &models.SubmitAuthorizeResult{
		Code: code,
	}, nil
}

func (s *authorizeService) validateSignin(ctx context.Context, project *entities.Project, req *models.SubmitAuthorizeRequest) (*entities.User, error) {
	userRepo := s.registry.UserRepository()
	email := strings.ToLower(req.Email) // Lowercase email before lookup
	user, err := userRepo.FindByProjectIDAndEmail(ctx, project.ID, email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if user == nil {
		log.WithContext(ctx).Errorw("User not found", "email", req.Email)
		return nil, errors.ErrInvalidUserCredentials
	}
	if !user.IsActive {
		log.WithContext(ctx).Errorw("User is not active", "user_id", user.ID)
		return nil, errors.ErrInvalidUserCredentials
	}
	if user.PasswordHash == nil {
		log.WithContext(ctx).Errorw("User has no password", "user_id", user.ID)
		return nil, errors.ErrInvalidUserCredentials
	}
	// Verify password
	if !passwordutil.IsValidPassword(req.Password, *user.PasswordHash) {
		log.WithContext(ctx).Errorw("Invalid password", "user_id", user.ID)
		return nil, errors.ErrInvalidUserCredentials
	}
	return user, nil
}

func (s *authorizeService) validateSignup(ctx context.Context, project *entities.Project, projectSetting *entities.ProjectSetting, req *models.SubmitAuthorizeRequest) error {
	if !projectSetting.IsSignupEnabled {
		log.WithContext(ctx).Errorw("Signup not enabled for project", "project_id", project.ID)
		return errors.ErrUnableToProcessSignup.WithDebugMessage("Signup not enabled for project")
	}
	if projectSetting.DefaultSignupRoleID == nil {
		log.WithContext(ctx).Errorw("Default signup role not set for project", "project_id", project.ID)
		return errors.ErrUnableToProcessSignup.WithDebugMessage("Default signup role not set for project")
	}
	if !projectSetting.IsAllowTempEmail {
		parsedEmail, err := disposable.ParseEmail(req.Email)
		if err != nil || parsedEmail.Disposable {
			log.WithContext(ctx).Errorw("Temporary email addresses are not allowed", "email", req.Email, "error", err)
			return errors.ErrDisposableEmail.WithError(err)
		}
	}
	// Check if user already exists
	userRepo := s.registry.UserRepository()
	existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, project.ID, req.Email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to check existing user", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if existingUser != nil {
		log.WithContext(ctx).Errorw("User already exists", "email", req.Email)
		return errors.ErrUnableToProcessSignup.WithDebugMessage(fmt.Sprintf("User already exists with email: %s", req.Email))
	}
	return nil
}

// processSignup handles the user creation and role assignment during signup.
// Signup at the system project (Roled Console) will create a new account for the user.
// Signup at other projects will use the project's account for the user.
func (s *authorizeService) processSignup(ctx context.Context,
	registry repositories.Registry,
	project *entities.Project,
	projectSetting *entities.ProjectSetting,
	redirectURI *entities.RedirectURI,
	req *models.SubmitAuthorizeRequest) (*entities.User, error) {

	email := strings.ToLower(req.Email) // Use lowercase email for user creation
	displayName := email[0:strings.Index(email, "@")]

	var err error
	var account *entities.Account

	// If it is a system project (Roled Console), create a new account for the user
	if project.IsSystem {
		account = &entities.Account{
			ID:       idutil.NewID(),
			Name:     displayName,
			IsActive: true,
			IsSystem: false,
		}
		accountRepo := registry.AccountRepository()
		err = accountRepo.Create(ctx, account)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create account", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
	} else {
		// Use project's account
		accountRepo := registry.AccountRepository()
		account, err = accountRepo.FindByID(ctx, project.AccountID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find project account", "account_id", project.AccountID, "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if account == nil {
			log.WithContext(ctx).Errorw("Project account not found", "account_id", project.AccountID)
			return nil, pkgerrors.ErrSystemError // This should not happen
		}
	}

	// Hash password
	passwordHash, err := passwordutil.HashPassword(req.Password)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to hash password", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Create user
	user := &entities.User{
		ID:           idutil.NewID(),
		AccountID:    account.ID,
		ProjectID:    project.ID,
		Email:        &email,
		PasswordHash: &passwordHash,
		DisplayName:  displayName,
		IsActive:     true,
	}
	userRepo := registry.UserRepository()
	err = userRepo.Create(ctx, user)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to create user", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	// Assign default role
	userRole := &entities.UserRole{
		UserID: user.ID,
		RoleID: *projectSetting.DefaultSignupRoleID,
	}
	userRoleRepo := registry.UserRoleRepository()
	err = userRoleRepo.Create(ctx, userRole)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to assign role", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// If project account is not a system account, create member record for this account
	if !account.IsSystem {
		member := &entities.Member{
			ID:        idutil.NewID(),
			AccountID: account.ID,
			UserID:    user.ID,
			IsAdmin:   true, // First user is admin by default
		}
		memberRepo := registry.MemberRepository()
		err = memberRepo.Create(ctx, member)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create member", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
	}

	if projectSetting.IsSignupVerifyEmail {
		// Publish verification email job to redis stream
		err = s.publishVerificationEmail(ctx, redirectURI.LoginURL, user, project)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to publish verification email", "error", err, "user_id", user.ID, "project_id", project.ID)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
	}

	return user, nil
}

func (s *authorizeService) publishVerificationEmail(ctx context.Context, loginURL *string, user *entities.User, project *entities.Project) error {
	payload := payloads.EmailPayload{
		Type:            constants.EmailPayloadTypeVerifyEmail,
		To:              *user.Email,
		From:            fmt.Sprintf(s.defaultConfig.Email.From, project.Name),
		Subject:         "Verify your email for " + project.Name,
		ProjectName:     project.Name,
		ProjectLogoURL:  project.LogoURL,
		ProjectIsSystem: project.IsSystem,
		LoginURL:        loginURL,
		UserID:          user.ID,
		IsSignup:        true,
	}
	contextFields := contextutil.GetFields(ctx, constants.RequestLoggerKeys)
	message := queues.Message{
		Payload: jsonutil.Stringify(payload),
		Context: jsonutil.Stringify(contextFields),
	}
	return s.emailPublisher.Publish(ctx, message)
}
