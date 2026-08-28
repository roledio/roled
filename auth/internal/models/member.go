package models

import (
	"time"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/pkg/models"
	"github.com/roledio/roled/pkg/utils/flashutil"
)

type GetMembersRequest struct {
	models.PageRequest
	AccountID      string     `query:"account_id"` // System account users can filter by account ID
	Search         string     `query:"search"`
	IsActive       *bool      `query:"is_active"`
	IsVerified     *bool      `query:"is_verified"`
	IsAdmin        *bool      `query:"is_admin"`
	CreatedAtSince *time.Time `query:"created_at_since"` // RFC 3339 format
	CreatedAtUntil *time.Time `query:"created_at_until"` // RFC 3339 format
}

type GetMembersResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	IsActive    bool      `json:"is_active"`
	IsVerified  bool      `json:"is_verified"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateMemberRequest struct {
	AccountID   string  `json:"account_id"` // System account users can create members for other accounts
	Email       string  `json:"email" validate:"required,email"`
	RedirectURI *string `json:"redirect_uri" validate:"omitempty,uri"` // Optional redirect URI for activation link
}

type CreateMemberResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsActive    bool   `json:"is_active"`
}

type RenderActivateMemberRequest struct {
	Token  string `uri:"token" validate:"required"`
	UserID *string
}

type RenderActivateMemberResponse struct {
	User     *entities.User
	Account  *entities.Account
	Project  *entities.Project
	LoginURL *string
}

type SubmitActivateMemberFlash struct {
	flashutil.FlashData
	DisplayName string
	UserID      string
	LoginURL    *string
	IsSuccess   bool
}

type SubmitActivateMemberRequest struct {
	Token                string `uri:"token" validate:"required"`
	DisplayName          string `form:"display_name" validate:"required,min=3,max=60"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,min=8,eqfield=Password"`
}

type SubmitActivateMemberResponse struct {
	UserID   string
	Account  *entities.Account
	Project  *entities.Project
	LoginURL *string
}

type DeleteMemberRequest struct {
	MemberID string `uri:"member_id" validate:"required"`
}

type UpdateMemberRequest struct {
	MemberID  string `uri:"member_id" validate:"required"`
	IsAdmin   *bool  `json:"is_admin"`                                      // nullable — only updated when provided
	AccountID string `json:"account_id" validate:"omitempty,min=10,max=50"` // system accounts can target a specific account
}

type UpdateMemberResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	IsAdmin   bool      `json:"is_admin"`
}

type GetMemberDetailsRequest struct {
	MemberID string `uri:"member_id" validate:"required"`
}

type GetMemberDetailsResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	IsActive    bool      `json:"is_active"`
	IsVerified  bool      `json:"is_verified"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateMemberTokenData struct {
	UserID   string
	LoginURL *string
}
