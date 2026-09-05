package authorize

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	autherrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/roledio/roled/auth/pkg/utils/randstrutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleOAuthStateExpiry = 10 * time.Minute
	googleProvider         = "google"
)

// InitiateGoogleOAuth initiates the Google OAuth flow
func (s *authorizeService) InitiateGoogleOAuth(ctx context.Context, req *models.GoogleOAuthRequest) (string, error) {
	// Validate the authorization request first
	project, _, _, err := s.validateAuthorizeRequest(ctx, &req.RenderAuthorizeRequest)
	if err != nil {
		return "", err
	}

	// Generate a cryptographically random state for Google OAuth
	googleState := randstrutil.RandomBase64(32)

	// Store the authorization context in Redis
	transaction := &models.GoogleOAuthTransaction{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		State:               req.State,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		IsSignup:            req.IsSignup,
		CreatedAt:           time.Now(),
	}

	// Store in Redis with expiry
	err = s.rediService.SetData(ctx, rediskeys.GoogleOAuthTransaction(googleState), transaction, googleOAuthStateExpiry)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to store Google OAuth transaction in Redis", "error", err)
		return "", pkgerrors.ErrSystemError.WithError(err)
	}

	// Create Google OAuth config
	googleOAuthConfig := s.getGoogleOAuthConfig()

	// Generate the authorization URL
	authURL := googleOAuthConfig.AuthCodeURL(googleState, oauth2.AccessTypeOffline)

	log.WithContext(ctx).Infow("Google OAuth initiated", "project_id", project.ID, "client_id", req.ClientID, "google_state", googleState)

	return authURL, nil
}

// HandleGoogleOAuthCallback handles the callback from Google OAuth
func (s *authorizeService) HandleGoogleOAuthCallback(ctx context.Context, req *models.GoogleOAuthCallbackRequest) (string, error) {
	// Retrieve the transaction from Redis
	cacheKey := rediskeys.GoogleOAuthTransaction(req.State)
	var transaction models.GoogleOAuthTransaction
	found, err := s.rediService.GetData(ctx, cacheKey, &transaction)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to retrieve Google OAuth transaction from Redis", "error", err)
		return "", pkgerrors.ErrSystemError.WithError(err)
	}
	if !found {
		log.WithContext(ctx).Errorw("Google OAuth transaction not found or expired", "state", req.State)
		return "", autherrors.ErrInvalidGoogleState
	}

	// Delete the transaction to prevent reuse (single-use)
	err = s.rediService.DeleteManyWithContext(ctx, []string{cacheKey})
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to delete Google OAuth transaction from Redis", "error", err)
	}

	// Create Google OAuth config
	googleOAuthConfig := s.getGoogleOAuthConfig()

	// Exchange the authorization code for tokens
	token, err := exchangeGoogleAuthCode(ctx, googleOAuthConfig, req.Code)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to exchange Google authorization code", "error", err)
		return "", autherrors.ErrGoogleTokenExchangeFailed
	}

	// Extract and validate the ID token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.WithContext(ctx).Errorw("No ID token in Google OAuth response")
		return "", autherrors.ErrGoogleIDTokenMissing
	}

	googleUserInfo, err := validateGoogleIDToken(ctx, idToken, s.defaultConfig.GoogleOAuth.ClientID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to validate Google ID token", "error", err)
		return "", err
	}

	// Validate the authorization request again (security check)
	renderReq := &models.RenderAuthorizeRequest{
		ClientID:            transaction.ClientID,
		RedirectURI:         transaction.RedirectURI,
		ResponseType:        "code",
		State:               transaction.State,
		CodeChallenge:       transaction.CodeChallenge,
		CodeChallengeMethod: transaction.CodeChallengeMethod,
		IsSignup:            transaction.IsSignup,
	}
	project, _, projectSetting, err := s.validateAuthorizeRequest(ctx, renderReq)
	if err != nil {
		return "", err
	}

	// Find or create user based on Google identity
	var redirectURL string
	err = s.registry.Tx(func(registry repositories.Registry) error {
		userIdentityRepo := registry.UserIdentityRepository()
		userRepo := registry.UserRepository()

		// Check if user identity already exists
		userIdentity, err := userIdentityRepo.FindByProviderAndProviderUserID(ctx, googleProvider, googleUserInfo.Sub)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find user identity", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		var user *entities.User
		if userIdentity != nil {
			// User identity exists, get the user
			user, err = userRepo.FindByID(ctx, userIdentity.UserID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find user by ID", "error", err)
				return pkgerrors.ErrSystemError.WithError(err)
			}
			if user == nil {
				log.WithContext(ctx).Errorw("User not found for existing identity", "user_identity", userIdentity)
				return pkgerrors.ErrSystemError
			}
			if !user.IsActive {
				log.WithContext(ctx).Errorw("User is not active", "user_id", user.ID)
				return autherrors.ErrInvalidUserCredentials
			}
		} else {
			// New user identity, create user
			user, err = s.createOrUpdateUserFromGoogle(ctx, registry, project, projectSetting, googleUserInfo)
			if err != nil {
				return err
			}
		}

		code, authCode := s.buildAuthCode(ctx, user, project, renderReq)
		authCodeRepo := registry.AuthCodeRepository()
		err = authCodeRepo.Create(ctx, authCode)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create auth code", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		log.WithContext(ctx).Infow("Google OAuth callback handled successfully", "user_id", user.ID, "project_id", project.ID)

		// Build the redirect URL with the authorization code
		redirectURL = fmt.Sprintf("%s?code=%s", transaction.RedirectURI, code)
		if transaction.State != "" {
			redirectURL += "&state=" + transaction.State
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return redirectURL, nil
}

// createOrUpdateUserFromGoogle creates or updates a user from Google OAuth information
func (s *authorizeService) createOrUpdateUserFromGoogle(ctx context.Context, registry repositories.Registry, project *entities.Project, projectSetting *entities.ProjectSetting, googleUserInfo *models.GoogleUserInfo) (*entities.User, error) {
	userRepo := registry.UserRepository()
	userIdentityRepo := registry.UserIdentityRepository()

	// Check if user exists with the same email
	existingUser, err := userRepo.FindByProjectIDAndEmail(ctx, project.ID, googleUserInfo.Email)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find user by email", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if existingUser != nil {
		// User exists with this email, link the Google identity
		userIdentity := &entities.UserIdentity{
			ID:             idutil.NewID(),
			UserID:         existingUser.ID,
			Provider:       googleProvider,
			ProviderUserID: googleUserInfo.Sub,
		}
		err = userIdentityRepo.Create(ctx, userIdentity)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create user identity", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		return existingUser, nil
	}

	if !projectSetting.IsSignupEnabled {
		log.WithContext(ctx).Errorw("Unable to continue signup with google oauth, signup is disabled for project", "project_id", project.ID)
		return nil, autherrors.ErrUnableToProcessSignup.WithDebugMessage("Signup not enabled for project")
	}

	// Create new user
	var account *entities.Account
	if project.IsSystem {
		// Create new account for system project
		account = &entities.Account{
			ID:       idutil.NewID(),
			Name:     googleUserInfo.Name,
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
			return nil, pkgerrors.ErrSystemError
		}
	}

	// Create user
	user := &entities.User{
		ID:          idutil.NewID(),
		AccountID:   account.ID,
		ProjectID:   project.ID,
		Email:       &googleUserInfo.Email,
		DisplayName: googleUserInfo.Name,
		AvatarURL:   &googleUserInfo.Picture,
		IsActive:    true,
	}

	if googleUserInfo.EmailVerified {
		now := time.Now()
		user.EmailVerifiedAt = &now
	}

	err = userRepo.Create(ctx, user)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to create user", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Create user identity
	userIdentity := &entities.UserIdentity{
		ID:             idutil.NewID(),
		UserID:         user.ID,
		Provider:       googleProvider,
		ProviderUserID: googleUserInfo.Sub,
	}
	err = userIdentityRepo.Create(ctx, userIdentity)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to create user identity", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	// Assign default role if signup is enabled
	if projectSetting.IsSignupEnabled && projectSetting.DefaultSignupRoleID != nil {
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
	}

	// If project account is not a system account, create member record
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

	return user, nil
}

func (s *authorizeService) getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.defaultConfig.GoogleOAuth.ClientID,
		ClientSecret: s.defaultConfig.GoogleOAuth.ClientSecret,
		RedirectURL:  s.defaultConfig.GoogleOAuth.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

type googleTokenExchanger func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error)

var exchangeGoogleAuthCode googleTokenExchanger = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(ctx, code)
}

type googleIDTokenValidator func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error)

var validateGoogleIDToken googleIDTokenValidator = defaultValidateGoogleIDToken

func defaultValidateGoogleIDToken(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
	return validateGoogleIDTokenWithIssuer(ctx, rawIDToken, clientID, "https://accounts.google.com")
}

// validateGoogleIDTokenWithIssuer validates the Google ID token and extracts user information
func validateGoogleIDTokenWithIssuer(ctx context.Context, rawIDToken, clientID, issuerURL string) (*models.GoogleUserInfo, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, autherrors.ErrGoogleIDTokenInvalid.WithError(err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, autherrors.ErrGoogleIDTokenInvalid.WithError(err)
	}

	// Extract claims
	var userInfo models.GoogleUserInfo
	err = idToken.Claims(&userInfo)
	if err != nil {
		return nil, autherrors.ErrGoogleIDTokenInvalid.WithError(err)
	}

	return &userInfo, nil
}
