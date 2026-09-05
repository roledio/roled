package authorize

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.openly.dev/pointy"
	"golang.org/x/oauth2"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	autherrors "github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
)

func createMockOIDCServer(t *testing.T, privateKey *rsa.PrivateKey) string {
	var server *httptest.Server

	nStr := base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes())
	eBytes := big.NewInt(int64(privateKey.E)).Bytes()
	eStr := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"alg": "RS256",
				"use": "sig",
				"kid": "test-key-id",
				"n":   nStr,
				"e":   eStr,
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":   server.URL,
			"jwks_uri": server.URL + "/keys",
		})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwk)
	})

	server = httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server.URL
}

func generateSignedJWT(t *testing.T, privateKey *rsa.PrivateKey, issuer, audience string, userInfo models.GoogleUserInfo, expiry time.Time) string {
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "test-key-id",
	}
	headerJSON, err := json.Marshal(header)
	assert.NoError(t, err)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	claims := map[string]any{
		"iss":            issuer,
		"aud":            audience,
		"sub":            userInfo.Sub,
		"email":          userInfo.Email,
		"email_verified": userInfo.EmailVerified,
		"name":           userInfo.Name,
		"picture":        userInfo.Picture,
		"exp":            expiry.Unix(),
		"iat":            time.Now().Unix(),
	}
	claimsJSON, err := json.Marshal(claims)
	assert.NoError(t, err)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hashed := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	assert.NoError(t, err)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigB64
}

// TestValidateGoogleIDToken_Success validates that a valid signed Google ID token is parsed correctly
func TestValidateGoogleIDToken_Success(t *testing.T) {
	ctx := context.Background()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey)
	clientID := "test-client-id"

	payload := models.GoogleUserInfo{
		Sub:           "google-sub-123",
		Email:         "testuser@gmail.com",
		EmailVerified: true,
		Name:          "Test User",
		Picture:       "https://example.com/photo.jpg",
	}

	idToken := generateSignedJWT(t, privateKey, issuerURL, clientID, payload, time.Now().Add(time.Hour))

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, clientID, issuerURL)

	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, "google-sub-123", userInfo.Sub)
	assert.Equal(t, "testuser@gmail.com", userInfo.Email)
	assert.True(t, userInfo.EmailVerified)
	assert.Equal(t, "Test User", userInfo.Name)
	assert.Equal(t, "https://example.com/photo.jpg", userInfo.Picture)
}

// TestValidateGoogleIDToken_InvalidSignature validates that an ID token with invalid signature is rejected
func TestValidateGoogleIDToken_InvalidSignature(t *testing.T) {
	ctx := context.Background()

	privateKey1, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKey2, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey1)
	clientID := "test-client-id"

	payload := models.GoogleUserInfo{
		Sub:   "google-sub-123",
		Email: "testuser@gmail.com",
	}

	// Sign with privateKey2 while server JWK only has privateKey1
	idToken := generateSignedJWT(t, privateKey2, issuerURL, clientID, payload, time.Now().Add(time.Hour))

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, clientID, issuerURL)

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.True(t, errors.Is(err, autherrors.ErrGoogleIDTokenInvalid))
}

// TestValidateGoogleIDToken_WrongAudience validates that an ID token with wrong audience is rejected
func TestValidateGoogleIDToken_WrongAudience(t *testing.T) {
	ctx := context.Background()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey)
	clientID := "test-client-id"

	payload := models.GoogleUserInfo{
		Sub:   "google-sub-123",
		Email: "testuser@gmail.com",
	}

	idToken := generateSignedJWT(t, privateKey, issuerURL, "wrong-client-id", payload, time.Now().Add(time.Hour))

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, clientID, issuerURL)

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.True(t, errors.Is(err, autherrors.ErrGoogleIDTokenInvalid))
}

// TestValidateGoogleIDToken_ExpiredToken validates that an expired ID token is rejected
func TestValidateGoogleIDToken_ExpiredToken(t *testing.T) {
	ctx := context.Background()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey)
	clientID := "test-client-id"

	payload := models.GoogleUserInfo{
		Sub:   "google-sub-123",
		Email: "testuser@gmail.com",
	}

	idToken := generateSignedJWT(t, privateKey, issuerURL, clientID, payload, time.Now().Add(-time.Hour))

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, clientID, issuerURL)

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.True(t, errors.Is(err, autherrors.ErrGoogleIDTokenInvalid))
}

// TestValidateGoogleIDToken_InvalidFormat validates that invalid token formats are rejected
func TestValidateGoogleIDToken_InvalidFormat(t *testing.T) {
	ctx := context.Background()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey)

	idToken := "invalid.token"

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, "test-client-id", issuerURL)

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.True(t, errors.Is(err, autherrors.ErrGoogleIDTokenInvalid))
}

// TestValidateGoogleIDToken_InvalidPayload validates that invalid payload is rejected
func TestValidateGoogleIDToken_InvalidPayload(t *testing.T) {
	ctx := context.Background()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	issuerURL := createMockOIDCServer(t, privateKey)

	idToken := "eyJhbGciOiJSUzI1NiJ9.invalid!!!.signature"

	userInfo, err := validateGoogleIDTokenWithIssuer(ctx, idToken, "test-client-id", issuerURL)

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.True(t, errors.Is(err, autherrors.ErrGoogleIDTokenInvalid))
}

// TestAuthorizeService_InitiateGoogleOAuth_Success validates initiating Google OAuth successfully
func TestAuthorizeService_InitiateGoogleOAuth_Success(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	mockRedisService.EXPECT().SetData(ctx, mock.MatchedBy(func(key string) bool {
		return len(key) > 0
	}), mock.AnythingOfType("*models.GoogleOAuthTransaction"), 10*time.Minute).Return(nil)

	defaultConfig := configs.DefaultConfig{}
	defaultConfig.GoogleOAuth.ClientID = "google-client-id"
	defaultConfig.GoogleOAuth.ClientSecret = "google-client-secret"
	defaultConfig.GoogleOAuth.RedirectURI = "http://localhost:8082/oauth/google/callback"

	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
	}

	authURL, err := service.InitiateGoogleOAuth(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "https://accounts.google.com/o/oauth2/auth")
	assert.Contains(t, authURL, "client_id=google-client-id")
}

// TestAuthorizeService_InitiateGoogleOAuth_ValidationFailed validates that invalid client returns error
func TestAuthorizeService_InitiateGoogleOAuth_ValidationFailed(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockClientRepo.EXPECT().FindByID(ctx, "invalid-client").Return(nil, nil)

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, mockRegistry, nil, nil)

	req := &models.GoogleOAuthRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:     "invalid-client",
			RedirectURI:  "http://example.com/callback",
			ResponseType: "code",
		},
	}

	authURL, err := service.InitiateGoogleOAuth(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, authURL)
	assert.Equal(t, autherrors.ErrInvalidClientID, err)
}

// TestAuthorizeService_InitiateGoogleOAuth_RedisError validates Redis failure handling
func TestAuthorizeService_InitiateGoogleOAuth_RedisError(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	mockRedisService.EXPECT().SetData(ctx, mock.Anything, mock.Anything, 10*time.Minute).Return(assert.AnError)

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthRequest{
		RenderAuthorizeRequest: models.RenderAuthorizeRequest{
			ClientID:            "test-client",
			RedirectURI:         "http://example.com/callback",
			ResponseType:        "code",
			CodeChallenge:       "test-challenge",
			CodeChallengeMethod: "S256",
			State:               "test-state",
		},
	}

	authURL, err := service.InitiateGoogleOAuth(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, authURL)
}

// TestAuthorizeService_HandleGoogleOAuthCallback_Success_ExistingIdentity tests callback for existing Google identity
func TestAuthorizeService_HandleGoogleOAuthCallback_Success_ExistingIdentity(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockUserIdentityRepo := repositorymocks.NewMockUserIdentityRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	state := "test-google-state"
	transaction := models.GoogleOAuthTransaction{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		State:               "original-state",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
		IsSignup:            false,
		CreatedAt:           time.Now(),
	}

	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	// Mock exchange and token verification
	origExchanger := exchangeGoogleAuthCode
	origValidator := validateGoogleIDToken
	t.Cleanup(func() {
		exchangeGoogleAuthCode = origExchanger
		validateGoogleIDToken = origValidator
	})

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		token := &oauth2.Token{
			AccessToken: "google-access-token",
		}
		return token.WithExtra(map[string]any{"id_token": "raw-id-token"}), nil
	}

	validateGoogleIDToken = func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
		return &models.GoogleUserInfo{
			Sub:           "google-sub-123",
			Email:         "testuser@gmail.com",
			EmailVerified: true,
			Name:          "Test User",
			Picture:       "https://example.com/avatar.jpg",
		}, nil
	}

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

	mockRegistry.EXPECT().UserIdentityRepository().Return(mockUserIdentityRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)

	userIdentity := &entities.UserIdentity{
		ID:             "identity-1",
		UserID:         "user-1",
		Provider:       "google",
		ProviderUserID: "google-sub-123",
	}
	mockUserIdentityRepo.EXPECT().FindByProviderAndProviderUserID(ctx, "google", "google-sub-123").Return(userIdentity, nil)

	user := &entities.User{
		ID:        "user-1",
		AccountID: "test-account",
		Email:     pointy.String("testuser@gmail.com"),
		IsActive:  true,
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-1").Return(user, nil)
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: state,
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.NoError(t, err)
	assert.Contains(t, redirectURL, "http://example.com/callback?code=")
	assert.Contains(t, redirectURL, "&state=original-state")
}

// TestAuthorizeService_HandleGoogleOAuthCallback_Success_NewUser_SystemProject tests callback for new user on system project
func TestAuthorizeService_HandleGoogleOAuthCallback_Success_NewUser_SystemProject(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockUserIdentityRepo := repositorymocks.NewMockUserIdentityRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockUserRoleRepo := repositorymocks.NewMockUserRoleRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	state := "test-google-state"
	transaction := models.GoogleOAuthTransaction{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		State:               "original-state",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now(),
	}

	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	origValidator := validateGoogleIDToken
	t.Cleanup(func() {
		exchangeGoogleAuthCode = origExchanger
		validateGoogleIDToken = origValidator
	})

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		token := &oauth2.Token{AccessToken: "google-access-token"}
		return token.WithExtra(map[string]any{"id_token": "raw-id-token"}), nil
	}

	validateGoogleIDToken = func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
		return &models.GoogleUserInfo{
			Sub:           "google-sub-456",
			Email:         "newgoogleuser@gmail.com",
			EmailVerified: true,
			Name:          "New Google User",
			Picture:       "https://example.com/avatar.jpg",
		}, nil
	}

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "system-project",
		AccountID: "system-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "system-project",
		IsActive: true,
		IsSystem: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "system-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "system-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "system-project", "http://example.com/callback").Return(redirectURI, nil)

	roleID := "default-role"
	projectSetting := &entities.ProjectSetting{
		ProjectID:           "system-project",
		IsSignupEnabled:     true,
		DefaultSignupRoleID: &roleID,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "system-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "system-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "system-account").Return(account, nil)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

	mockRegistry.EXPECT().UserIdentityRepository().Return(mockUserIdentityRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().UserRoleRepository().Return(mockUserRoleRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)

	// User identity not found
	mockUserIdentityRepo.EXPECT().FindByProviderAndProviderUserID(ctx, "google", "google-sub-456").Return(nil, nil)
	// User by email not found
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "system-project", "newgoogleuser@gmail.com").Return(nil, nil)
	// Create account for system project
	mockAccountRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.Account")).Return(nil)
	// Create user
	mockUserRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.User")).Return(nil)
	// Create user identity
	mockUserIdentityRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserIdentity")).Return(nil)
	// Assign default role
	mockUserRoleRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserRole")).Return(nil)
	// Create member
	mockMemberRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.Member")).Return(nil)
	// Create auth code
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: state,
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.NoError(t, err)
	assert.Contains(t, redirectURL, "http://example.com/callback?code=")
	assert.Contains(t, redirectURL, "&state=original-state")
}

// TestAuthorizeService_HandleGoogleOAuthCallback_Success_NewUser_NonSystemProject tests callback for new user on non-system project
func TestAuthorizeService_HandleGoogleOAuthCallback_Success_NewUser_NonSystemProject(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockUserIdentityRepo := repositorymocks.NewMockUserIdentityRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockUserRoleRepo := repositorymocks.NewMockUserRoleRepository(t)
	mockMemberRepo := repositorymocks.NewMockMemberRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	state := "test-google-state"
	transaction := models.GoogleOAuthTransaction{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		State:               "original-state",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
		IsSignup:            true,
		CreatedAt:           time.Now(),
	}

	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	origValidator := validateGoogleIDToken
	t.Cleanup(func() {
		exchangeGoogleAuthCode = origExchanger
		validateGoogleIDToken = origValidator
	})

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		token := &oauth2.Token{AccessToken: "google-access-token"}
		return token.WithExtra(map[string]any{"id_token": "raw-id-token"}), nil
	}

	validateGoogleIDToken = func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
		return &models.GoogleUserInfo{
			Sub:           "google-sub-non-system",
			Email:         "nonsystemuser@gmail.com",
			EmailVerified: true,
			Name:          "Non System User",
			Picture:       "https://example.com/avatar.jpg",
		}, nil
	}

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "project-1",
		AccountID: "account-1",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:        "project-1",
		AccountID: "account-1",
		IsActive:  true,
		IsSystem:  false,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "project-1").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "project-1",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "project-1", "http://example.com/callback").Return(redirectURI, nil)

	roleID := "default-role"
	projectSetting := &entities.ProjectSetting{
		ProjectID:           "project-1",
		IsSignupEnabled:     true,
		DefaultSignupRoleID: &roleID,
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "project-1").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "account-1",
		IsActive: true,
		IsSystem: false,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "account-1").Return(account, nil)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

	mockRegistry.EXPECT().UserIdentityRepository().Return(mockUserIdentityRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().UserRoleRepository().Return(mockUserRoleRepo)
	mockRegistry.EXPECT().MemberRepository().Return(mockMemberRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)

	// User identity not found
	mockUserIdentityRepo.EXPECT().FindByProviderAndProviderUserID(ctx, "google", "google-sub-non-system").Return(nil, nil)
	// User by email not found
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "project-1", "nonsystemuser@gmail.com").Return(nil, nil)
	// Find project account
	mockAccountRepo.EXPECT().FindByID(ctx, "account-1").Return(account, nil)
	// Create user
	mockUserRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.User")).Return(nil)
	// Create user identity
	mockUserIdentityRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserIdentity")).Return(nil)
	// Assign default role
	mockUserRoleRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserRole")).Return(nil)
	// Create member
	mockMemberRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.Member")).Return(nil)
	// Create auth code
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: state,
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.NoError(t, err)
	assert.Contains(t, redirectURL, "http://example.com/callback?code=")
	assert.Contains(t, redirectURL, "&state=original-state")
}

// TestAuthorizeService_HandleGoogleOAuthCallback_Success_UserExistsByEmail tests linking Google identity to existing email user
func TestAuthorizeService_HandleGoogleOAuthCallback_Success_UserExistsByEmail(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockUserIdentityRepo := repositorymocks.NewMockUserIdentityRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockAuthCodeRepo := repositorymocks.NewMockAuthCodeRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	state := "test-google-state"
	transaction := models.GoogleOAuthTransaction{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		State:               "original-state",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
		CreatedAt:           time.Now(),
	}

	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	origValidator := validateGoogleIDToken
	t.Cleanup(func() {
		exchangeGoogleAuthCode = origExchanger
		validateGoogleIDToken = origValidator
	})

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		token := &oauth2.Token{AccessToken: "google-access-token"}
		return token.WithExtra(map[string]any{"id_token": "raw-id-token"}), nil
	}

	validateGoogleIDToken = func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
		return &models.GoogleUserInfo{
			Sub:           "google-sub-789",
			Email:         "existinguser@gmail.com",
			EmailVerified: true,
			Name:          "Existing User",
		}, nil
	}

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

	mockRegistry.EXPECT().UserIdentityRepository().Return(mockUserIdentityRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)
	mockRegistry.EXPECT().AuthCodeRepository().Return(mockAuthCodeRepo)

	// User identity not found
	mockUserIdentityRepo.EXPECT().FindByProviderAndProviderUserID(ctx, "google", "google-sub-789").Return(nil, nil)

	// User found by email
	existingUser := &entities.User{
		ID:        "existing-user-id",
		AccountID: "test-account",
		Email:     pointy.String("existinguser@gmail.com"),
		IsActive:  true,
	}
	mockUserRepo.EXPECT().FindByProjectIDAndEmail(ctx, "test-project", "existinguser@gmail.com").Return(existingUser, nil)

	// Link identity
	mockUserIdentityRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.UserIdentity")).Return(nil)
	// Create auth code
	mockAuthCodeRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entities.AuthCode")).Return(nil)

	defaultConfig := configs.DefaultConfig{}
	defaultConfig.JWT.AuthCodeExpiryDuration = "1m"
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: state,
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.NoError(t, err)
	assert.Contains(t, redirectURL, "http://example.com/callback?code=")
	assert.Contains(t, redirectURL, "&state=original-state")
}

// TestAuthorizeService_HandleGoogleOAuthCallback_TransactionNotFound tests callback when transaction is missing or expired in Redis
func TestAuthorizeService_HandleGoogleOAuthCallback_TransactionNotFound(t *testing.T) {
	ctx := context.Background()

	mockRedisService := servicemocks.NewMockRedisService(t)
	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).Return(false, nil)

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, nil, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: "expired-state",
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, redirectURL)
	assert.Equal(t, autherrors.ErrInvalidGoogleState, err)
}

// TestAuthorizeService_HandleGoogleOAuthCallback_TokenExchangeFailed tests error when Google token exchange fails
func TestAuthorizeService_HandleGoogleOAuthCallback_TokenExchangeFailed(t *testing.T) {
	ctx := context.Background()

	mockRedisService := servicemocks.NewMockRedisService(t)

	transaction := models.GoogleOAuthTransaction{
		ClientID:    "test-client",
		RedirectURI: "http://example.com/callback",
		State:       "original-state",
	}
	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	t.Cleanup(func() { exchangeGoogleAuthCode = origExchanger })

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		return nil, errors.New("google exchange error")
	}

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, nil, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "bad-code",
		State: "test-state",
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, redirectURL)
	assert.Equal(t, autherrors.ErrGoogleTokenExchangeFailed, err)
}

// TestAuthorizeService_HandleGoogleOAuthCallback_IDTokenMissing tests error when ID token is missing from token response
func TestAuthorizeService_HandleGoogleOAuthCallback_IDTokenMissing(t *testing.T) {
	ctx := context.Background()

	mockRedisService := servicemocks.NewMockRedisService(t)

	transaction := models.GoogleOAuthTransaction{
		ClientID:    "test-client",
		RedirectURI: "http://example.com/callback",
		State:       "original-state",
	}
	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	t.Cleanup(func() { exchangeGoogleAuthCode = origExchanger })

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		// Token with no id_token extra
		return &oauth2.Token{AccessToken: "access-token"}, nil
	}

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, nil, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "code",
		State: "test-state",
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, redirectURL)
	assert.Equal(t, autherrors.ErrGoogleIDTokenMissing, err)
}

// TestAuthorizeService_HandleGoogleOAuthCallback_UserInactive tests error when user is not active
func TestAuthorizeService_HandleGoogleOAuthCallback_UserInactive(t *testing.T) {
	ctx := context.Background()

	mockRegistry := repositorymocks.NewMockRegistry(t)
	mockClientRepo := repositorymocks.NewMockClientRepository(t)
	mockProjectRepo := repositorymocks.NewMockProjectRepository(t)
	mockProjectSettingRepo := repositorymocks.NewMockProjectSettingRepository(t)
	mockAccountRepo := repositorymocks.NewMockAccountRepository(t)
	mockRedirectURIRepo := repositorymocks.NewMockRedirectURIRepository(t)
	mockUserIdentityRepo := repositorymocks.NewMockUserIdentityRepository(t)
	mockUserRepo := repositorymocks.NewMockUserRepository(t)
	mockRedisService := servicemocks.NewMockRedisService(t)

	state := "test-google-state"
	transaction := models.GoogleOAuthTransaction{
		ClientID:            "test-client",
		RedirectURI:         "http://example.com/callback",
		State:               "original-state",
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
	}

	mockRedisService.EXPECT().GetData(ctx, mock.Anything, mock.AnythingOfType("*models.GoogleOAuthTransaction")).RunAndReturn(
		func(ctx context.Context, key string, dest any) (bool, error) {
			ptr := dest.(*models.GoogleOAuthTransaction)
			*ptr = transaction
			return true, nil
		},
	)
	mockRedisService.EXPECT().DeleteManyWithContext(ctx, mock.Anything).Return(nil)

	origExchanger := exchangeGoogleAuthCode
	origValidator := validateGoogleIDToken
	t.Cleanup(func() {
		exchangeGoogleAuthCode = origExchanger
		validateGoogleIDToken = origValidator
	})

	exchangeGoogleAuthCode = func(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
		token := &oauth2.Token{AccessToken: "google-access-token"}
		return token.WithExtra(map[string]any{"id_token": "raw-id-token"}), nil
	}

	validateGoogleIDToken = func(ctx context.Context, rawIDToken, clientID string) (*models.GoogleUserInfo, error) {
		return &models.GoogleUserInfo{
			Sub:   "google-sub-123",
			Email: "inactive@gmail.com",
		}, nil
	}

	mockRegistry.EXPECT().ClientRepository().Return(mockClientRepo)
	mockRegistry.EXPECT().ProjectRepository().Return(mockProjectRepo)
	mockRegistry.EXPECT().ProjectSettingRepository().Return(mockProjectSettingRepo)
	mockRegistry.EXPECT().AccountRepository().Return(mockAccountRepo)
	mockRegistry.EXPECT().RedirectURIRepository().Return(mockRedirectURIRepo)

	client := &entities.Client{
		ID:        "test-client",
		ProjectID: "test-project",
		AccountID: "test-account",
		IsActive:  true,
	}
	mockClientRepo.EXPECT().FindByID(ctx, "test-client").Return(client, nil)

	project := &entities.Project{
		ID:       "test-project",
		IsActive: true,
	}
	mockProjectRepo.EXPECT().FindByID(ctx, "test-project").Return(project, nil)

	redirectURI := &entities.RedirectURI{
		ProjectID:   "test-project",
		RedirectURI: "http://example.com/callback",
	}
	mockRedirectURIRepo.EXPECT().FindByProjectIDAndRedirectURI(ctx, "test-project", "http://example.com/callback").Return(redirectURI, nil)

	projectSetting := &entities.ProjectSetting{
		ProjectID: "test-project",
	}
	mockProjectSettingRepo.EXPECT().FindByProjectID(ctx, "test-project").Return(projectSetting, nil)

	account := &entities.Account{
		ID:       "test-account",
		IsActive: true,
	}
	mockAccountRepo.EXPECT().FindByID(ctx, "test-account").Return(account, nil)

	mockRegistry.EXPECT().Tx(mock.AnythingOfType("func(repositories.Registry) error")).RunAndReturn(func(fn func(repositories.Registry) error) error {
		return fn(mockRegistry)
	})

	mockRegistry.EXPECT().UserIdentityRepository().Return(mockUserIdentityRepo)
	mockRegistry.EXPECT().UserRepository().Return(mockUserRepo)

	userIdentity := &entities.UserIdentity{
		ID:             "identity-1",
		UserID:         "user-1",
		Provider:       "google",
		ProviderUserID: "google-sub-123",
	}
	mockUserIdentityRepo.EXPECT().FindByProviderAndProviderUserID(ctx, "google", "google-sub-123").Return(userIdentity, nil)

	user := &entities.User{
		ID:        "user-1",
		AccountID: "test-account",
		Email:     pointy.String("inactive@gmail.com"),
		IsActive:  false, // Inactive user
	}
	mockUserRepo.EXPECT().FindByID(ctx, "user-1").Return(user, nil)

	defaultConfig := configs.DefaultConfig{}
	service := NewAuthorizeService(&defaultConfig, mockRegistry, mockRedisService, nil)

	req := &models.GoogleOAuthCallbackRequest{
		Code:  "google-auth-code",
		State: state,
	}

	redirectURL, err := service.HandleGoogleOAuthCallback(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, redirectURL)
	assert.Equal(t, autherrors.ErrInvalidUserCredentials, err)
}
