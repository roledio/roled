package models

import (
	"time"

	"github.com/roledio/roled/auth/pkg/models"
)

type GetPermissionsRequest struct {
	models.PageRequest
	ProjectID string `uri:"project_id" validate:"required"`
	Search    string `query:"search"`
	IsDefault *bool  `query:"is_default"`
}

type PermissionDetails struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Description  *string   `json:"description"`
	IsDefault    bool      `json:"is_default"`
	ResourceID   string    `json:"resource_id"`
	ResourceName string    `json:"resource_name"`
	ResourceCode string    `json:"resource_code"`
}
