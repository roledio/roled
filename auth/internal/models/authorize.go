package models

import (
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/pkg/utils/flashutil"
)

type RenderAuthorizeRequest struct {
	ClientID            string `form:"client_id" query:"client_id" validate:"required"`
	RedirectURI         string `form:"redirect_uri" query:"redirect_uri" validate:"required,uri"`
	ResponseType        string `form:"response_type" query:"response_type" validate:"required,oneof=code"` // Implicit flow with response_type=token is not supported
	State               string `form:"state" query:"state"`
	CodeChallenge       string `form:"code_challenge" query:"code_challenge" validate:"required,base64rawurl"`
	CodeChallengeMethod string `form:"code_challenge_method" query:"code_challenge_method" validate:"required,oneof=S256"` // Method "plain" (no transformation) is not supported, only "S256"

	IsSignup bool `form:"is_signup"` // Always read from form, query is used to render the form initially
}

type RenderAuthorizeResult struct {
	Project        *entities.Project
	ProjectSetting *entities.ProjectSetting
}

type SubmitAuthorizeRequest struct {
	RenderAuthorizeRequest
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"omitempty,required_if=IsSignup true,eqfield=Password"`
}

type SubmitAuthorizeResult struct {
	Code string
}

type SubmitAuthorizeFlash struct {
	flashutil.FlashData
	Email    string
	IsSignup bool
}

type EmailVerifyTokenData struct {
	UserID   string
	LoginURL *string
}

type GoogleOAuthRequest struct {
	RenderAuthorizeRequest
}
