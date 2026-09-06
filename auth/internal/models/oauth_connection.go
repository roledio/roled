package models

import (
	"time"
)

type GetOAuthConnectionsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
}

type OAuthConnectionDetails struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ProjectID string    `json:"project_id"`
	Provider  string    `json:"provider"`
	Enabled   bool      `json:"enabled"`
}
