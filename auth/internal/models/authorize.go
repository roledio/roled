package models

import (
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/pkg/utils/flashutil"
)

type RenderAuthorizeRequest struct {
	ClientID            string `query:"client_id" validate:"required"`
	RedirectURI         string `query:"redirect_uri" validate:"required,uri"`
	ResponseType        string `query:"response_type" validate:"required,oneof=code"` // Implicit flow with response_type=token is not supported
	State               string `query:"state"`
	CodeChallenge       string `query:"code_challenge" validate:"required,base64rawurl"`
	CodeChallengeMethod string `query:"code_challenge_method" validate:"required,oneof=S256"` // Method "plain" (no transformation) is not supported, only "S256"
	IsSignup            bool   `query:"signup"`
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
	IsSignup             bool   `form:"is_signup"`
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
