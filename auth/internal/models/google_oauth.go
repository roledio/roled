package models

import "time"

// GoogleOAuthTransaction stores the authorization context in Redis during Google OAuth flow
type GoogleOAuthTransaction struct {
	ClientID            string    `json:"client_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	State               string    `json:"state"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
	IsSignup            bool      `json:"is_signup"`
	CreatedAt           time.Time `json:"created_at"`
}

// GoogleOAuthCallbackRequest represents the request from Google OAuth callback
type GoogleOAuthCallbackRequest struct {
	Code  string `query:"code" validate:"required"`
	State string `query:"state" validate:"required"`
}

// GoogleUserInfo represents the user information from Google ID token
type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
