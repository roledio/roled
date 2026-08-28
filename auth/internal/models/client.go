package models

import (
	"time"

	"github.com/roledio/roled/pkg/models"
)

type GetClientsRequest struct {
	models.PageRequest
	ProjectID      string     `uri:"project_id" validate:"required"`
	Search         string     `query:"search"`
	IsActive       *bool      `query:"is_active"`
	CreatedAtSince *time.Time `query:"created_at_since"` // RFC 3339 format
	CreatedAtUntil *time.Time `query:"created_at_until"` // RFC 3339 format
}

type ClientDetails struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Description   *string   `json:"description"`
	IsDefault     bool      `json:"is_default"`
	IsActive      bool      `json:"is_active"`
	Secret        string    `json:"secret,omitempty"`
	PermissionIDs []string  `json:"permission_ids,omitempty"`
}

type GetClientDetailsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	ClientID  string `uri:"client_id" validate:"required"`
}

type CreateClientRequest struct {
	ProjectID     string   `uri:"project_id" validate:"required"`
	Name          string   `json:"name" validate:"required,notblank,max=50"`
	Description   *string  `json:"description" validate:"omitempty,max=200"`
	PermissionIDs []string `json:"permission_ids"`
}

type ClientPermission struct {
	ID             string `json:"id"`
	ResourceName   string `json:"resource_name"`
	PermissionName string `json:"permission_name"`
}

type UpdateClientRequest struct {
	ProjectID     string   `uri:"project_id" validate:"required"`
	ClientID      string   `uri:"client_id" validate:"required"`
	Name          string   `json:"name" validate:"required,notblank,max=50"`
	Description   *string  `json:"description" validate:"omitempty,max=200"`
	IsActive      *bool    `json:"is_active" validate:"required"`
	PermissionIDs []string `json:"permission_ids"`
}

type ClientDetailsAndPermissions struct {
	ClientDetails
	Permissions []ClientPermission `json:"permissions"`
}

type DeleteClientRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	ClientID  string `uri:"client_id" validate:"required"`
}
