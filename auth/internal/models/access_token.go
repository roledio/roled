package models

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/pkg/utils/jwtutil"
)

type ExchangeTokenRequest struct {
	// Client credentials flow must use basic auth header
	Authorization     string `header:"Authorization" validate:"required_if=GrantType client_credentials"`
	GrantType         string `form:"grant_type" json:"grant_type" validate:"required,oneof=authorization_code refresh_token client_credentials"`
	ClientID          string `form:"client_id" json:"client_id" validate:"required_unless=GrantType client_credentials"`
	ClientSecret      string // Will be extracted from Authorization header for client credentials flow
	RefreshToken      string `form:"refresh_token" json:"refresh_token" validate:"required_if=GrantType refresh_token"`
	AuthorizationCode string `form:"authorization_code" json:"authorization_code" validate:"required_if=GrantType authorization_code"`
	RedirectURI       string `form:"redirect_uri" json:"redirect_uri" validate:"required_if=GrantType authorization_code"`
	State             string `form:"state" json:"state"`
	CodeVerifier      string `form:"code_verifier" json:"code_verifier" validate:"required_if=GrantType authorization_code"`
}

func (r *ExchangeTokenRequest) ValidateAuthorization(ctx context.Context) error {
	// Parse basic auth for client credentials flow
	if r.GrantType == constants.GrantTypeClientCredentials {
		if !strings.HasPrefix(r.Authorization, "Basic ") {
			err := errors.New("invalid authentication scheme in Authorization header, expected Basic")
			log.WithContext(ctx).Errorw("Authorization header must start with Basic for client credentials flow", "authorization_header", r.Authorization)
			return err
		}
		encoded := r.Authorization[len("Basic "):]
		decoded, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to decode basic auth credentials", "error", err)
			return err
		}
		credentials := string(decoded)
		split := strings.Split(credentials, ":")
		if len(split) != 2 {
			err := errors.New("invalid Basic auth credentials format")
			log.WithContext(ctx).Errorw("Invalid Basic auth credentials format", "credentials", credentials)
			return err
		}
		r.ClientID = split[0]
		r.ClientSecret = split[1]
	}
	return nil
}

type ExchangeTokenResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"` // in seconds
	TokenType             string `json:"token_type"` // always "bearer"
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in,omitempty"`
}

type AccessTokenProject struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

type AccessTokenClient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccessTokenUser struct {
	ID             string  `json:"id"`
	DisplayName    string  `json:"display_name"`
	Email          *string `json:"email"`
	ExternalUserID *string `json:"external_user_id"`
	AvatarURL      *string `json:"avatar_url"`
}

type AccessTokenRole struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AccessTokenDetails struct {
	ID          string             `json:"id"`
	IssuedAt    time.Time          `json:"issued_at"`
	ExpiresAt   time.Time          `json:"expires_at"`
	Project     AccessTokenProject `json:"project"`
	Client      AccessTokenClient  `json:"client"`
	User        *AccessTokenUser   `json:"user,omitempty"`
	Role        *AccessTokenRole   `json:"role,omitempty"`
	Permissions []string           `json:"permissions"`
}

type RevokeCurrentTokenRequest struct {
	Authorization string `header:"Authorization"` // Optional access token
	ClientID      string `json:"client_id" validate:"required"`
	RefreshToken  string `json:"refresh_token" validate:"required"`
	JWTClaims     *JWTClaims
}

func (r *RevokeCurrentTokenRequest) ParseJWT(ctx context.Context, defaultConfig *configs.DefaultConfig) {
	if !strings.HasPrefix(r.Authorization, "Bearer ") {
		log.WithContext(ctx).Warnw("Authorization header does not start with Bearer for token revocation, skipping JWT parsing", "authorization_header", r.Authorization)
		return
	}
	token := r.Authorization[len("Bearer "):]
	trim := strings.TrimSpace(token)
	if trim == "" {
		log.WithContext(ctx).Warn("No JWT provided in Authorization header for token revocation")
		return
	}
	var claims JWTClaims
	err := jwtutil.ParseToken(trim, defaultConfig.JWT.SigningKey, &claims)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to parse JWT from Authorization header for token revocation", "error", err)
		return
	}
	r.JWTClaims = &claims
}
