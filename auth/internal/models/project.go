package models

import (
	"time"

	"github.com/roledio/roled/pkg/models"
)

type GetProjectsRequest struct {
	models.PageRequest
	Search         string     `query:"search"`
	IsActive       *bool      `query:"is_active"`
	CreatedAtSince *time.Time `query:"created_at_since"` // RFC 3339 format
	CreatedAtUntil *time.Time `query:"created_at_until"` // RFC 3339 format
}

type GetProjectsResponse struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	LogoURL     *string   `json:"logo_url"`
	IsActive    bool      `json:"is_active"`
}

type GetProjectDetailsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
}

type ProjectDetails struct {
	ID           string        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Name         string        `json:"name"`
	Description  *string       `json:"description"`
	LogoURL      *string       `json:"logo_url"`
	RedirectURIs []RedirectURI `json:"redirect_uris"`
	IsActive     bool          `json:"is_active"`
}

type RedirectURI struct {
	RedirectURI string `json:"redirect_uri" validate:"required,uri"`
	LoginURL    string `json:"login_url" validate:"omitempty,url"`
}

type CreateProjectRequest struct {
	Name         string        `json:"name" validate:"required,notblank,max=50"`
	Description  *string       `json:"description" validate:"omitempty,max=400"`
	LogoURL      *string       `json:"logo_url" validate:"omitempty,url"`
	RedirectURIs []RedirectURI `json:"redirect_uris" validate:"omitempty,dive"`
}

type UpdateProjectRequest struct {
	ProjectID    string        `uri:"project_id" validate:"required"`
	Name         string        `json:"name" validate:"required,notblank,max=50"`
	Description  *string       `json:"description" validate:"omitempty,max=400"`
	LogoURL      *string       `json:"logo_url" validate:"omitempty,url"`
	RedirectURIs []RedirectURI `json:"redirect_uris" validate:"omitempty,dive"`
	IsActive     *bool         `json:"is_active" validate:"required"`
}

type DeleteProjectRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	Name      string `json:"name" validate:"required"`
}
