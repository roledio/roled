package models

import (
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/pkg/utils/flashutil"
)

type RenderForgotPasswordRequest struct {
	ClientID    string `query:"client_id" validate:"required"`
	RedirectURI string `query:"redirect_uri" validate:"omitempty,uri"`
}

type RenderForgotPasswordResult struct {
	Project  *entities.Project
	LoginURL *string
}

type SubmitForgotPasswordRequest struct {
	ClientID    string `form:"client_id" validate:"required"`
	RedirectURI string `form:"redirect_uri" validate:"omitempty,uri"`
	Email       string `form:"email" validate:"required,email"`
}

type SubmitForgotPasswordResult struct {
	Project *entities.Project
}

type SubmitForgotPasswordFlash struct {
	flashutil.FlashData
	Email     string
	IsSuccess bool
}

type ResetPasswordTokenData struct {
	UserID   string
	LoginURL *string
}

type RenderResetPasswordRequest struct {
	Token     string `uri:"token" validate:"required"`
	ProjectID *string
}

type RenderResetPasswordResult struct {
	Project *entities.Project
}

type SubmitResetPasswordFlash struct {
	flashutil.FlashData
	IsSuccess bool
	LoginURL  *string
	ProjectID string
}

type SubmitResetPasswordRequest struct {
	Token                string `uri:"token" validate:"required"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,min=8,eqfield=Password"`
}

type SubmitResetPasswordResult struct {
	Project  *entities.Project
	LoginURL *string
}
