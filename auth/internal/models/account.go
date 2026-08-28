package models

import (
	"time"

	"github.com/roledio/roled/pkg/models"
)

type GetAccountDetailsRequest struct {
	AccountID string `uri:"account_id" validate:"required"`
}

type GetAccountDetailsResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GetAccountsRequest struct {
	models.PageRequest
	Search         string     `query:"search"`
	IsActive       *bool      `query:"is_active"`
	CreatedAtSince *time.Time `query:"created_at_since"` // RFC 3339 format
	CreatedAtUntil *time.Time `query:"created_at_until"` // RFC 3339 format
}

type GetAccountsResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateAccountRequest struct {
	AccountID   string `uri:"account_id" validate:"required"`
	Name        string `json:"name" validate:"min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	IsActive    bool   `json:"is_active"`
}

type UpdateAccountResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeleteAccountRequest struct {
	AccountID string  `uri:"account_id" validate:"required"`
	Password  *string `json:"password"`
}
