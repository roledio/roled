package models

import (
	"time"

	"github.com/roledio/roled/auth/pkg/models"
)

type GetUsersRequest struct {
	models.PageRequest
	ProjectID string  `uri:"project_id" validate:"required"`
	Search    string  `query:"search"`
	IsActive  *bool   `query:"is_active"`
	RoleID    *string `query:"role_id"`
}

type GetUserDetailsRequest struct {
	ProjectID          string `uri:"project_id" validate:"required"`
	UserID             string `uri:"user_id" validate:"required"`
	IncludePermissions bool   `query:"include_permissions"`
}

type UserDetails struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Email           *string   `json:"email"`
	ExternalUserID  *string   `json:"external_user_id"`
	DisplayName     string    `json:"display_name"`
	AvatarURL       *string   `json:"avatar_url"`
	IsActive        bool      `json:"is_active"`
	IsEmailVerified bool      `json:"is_email_verified"`
	RoleID          string    `json:"role_id,omitempty"`
	RoleName        string    `json:"role_name,omitempty"`
	Permissions     []string  `json:"permissions,omitempty"`
}

type CreateUserRequest struct {
	ProjectID      string `uri:"project_id" validate:"required"`
	DisplayName    string `json:"display_name" validate:"required,notblank,max=50"`
	AvatarURL      string `json:"avatar_url" validate:"omitempty,url"`
	ExternalUserID string `json:"external_user_id" validate:"required_without=Email,notblank,omitempty,max=100"`                    // Required if Email is not provided
	Email          string `json:"email" validate:"omitempty,email,max=100"`                                                         // Optional if ExternalUserID is provided
	Password       string `json:"password" validate:"required_without=ExternalUserID,required_with=Email,notblank,omitempty,min=8"` // Required if ExternalUserID is not provided
	RoleID         string `json:"role_id"`
}

type UpdateUserRequest struct {
	ProjectID      string `uri:"project_id" validate:"required"`
	UserID         string `uri:"user_id" validate:"required"`
	DisplayName    string `json:"display_name" validate:"required,notblank,max=50"`
	AvatarURL      string `json:"avatar_url" validate:"omitempty,url"`
	ExternalUserID string `json:"external_user_id" validate:"required_without=Email,notblank,omitempty,max=100"` // Required if Email is not provided
	Email          string `json:"email" validate:"omitempty,email,max=100"`                                      // Optional if ExternalUserID is provided
	Password       string `json:"password" validate:"omitempty,notblank,min=8"`                                  // Password is optional in update
	IsActive       *bool  `json:"is_active" validate:"required"`
	RoleID         string `json:"role_id"`
}

type DeleteUserRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	UserID    string `uri:"user_id" validate:"required"`
}

type UpdateCurrentUserRequest struct {
	DisplayName string `json:"display_name" validate:"required,notblank,max=50"`
	AvatarURL   string `json:"avatar_url" validate:"omitempty,url"`
	Email       string `json:"email" validate:"required,email,max=100"`
	Password    string `json:"password" validate:"omitempty,notblank,min=8"` // Password is optional in update
}

type GetExternalUserDetailsRequest struct {
	ProjectID          string `uri:"project_id" validate:"required"`
	ExternalUserID     string `uri:"external_user_id" validate:"required"`
	IncludePermissions bool   `query:"include_permissions"`
}

type ResendVerificationEmailRequest struct {
	ProjectID   string  `uri:"project_id" validate:"required"`
	UserID      string  `uri:"user_id" validate:"required"`
	RedirectURI *string `json:"redirect_uri" validate:"omitempty,uri"` // Optional redirect URI for login link
}

type RequestPasswordResetRequest struct {
	ProjectID   string  `uri:"project_id" validate:"required"`
	UserID      string  `uri:"user_id" validate:"required"`
	RedirectURI *string `json:"redirect_uri" validate:"omitempty,uri"` // Optional redirect URI for login link
}
