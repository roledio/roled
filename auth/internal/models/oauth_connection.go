package models

import (
	"time"
)

type GetOAuthConnectionsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
}

type GetOAuthConnectionRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	Provider  string `uri:"provider" validate:"required,oneof=google"`
}

type OAuthConnectionDetails struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ProjectID      string    `json:"project_id"`
	Provider       string    `json:"provider"`
	CredentialType string    `json:"credential_type"`
	Enabled        bool      `json:"enabled"`
	ClientID       *string   `json:"client_id,omitempty"`
	ClientSecret   *string   `json:"client_secret,omitempty"`
	Scopes         []string  `json:"scopes,omitempty"`
}

type CreateOAuthConnectionRequest struct {
	ProjectID      string   `uri:"project_id" validate:"required"`
	Provider       string   `uri:"provider" validate:"required,oneof=google"`
	CredentialType string   `json:"credential_type" validate:"required,oneof=default custom"`
	ClientID       string   `json:"client_id" validate:"required_if=CredentialType custom,notblank,max=512"`
	ClientSecret   string   `json:"client_secret" validate:"required_if=CredentialType custom,notblank,max=512"`
	Scopes         []string `json:"scopes" validate:"required_if=CredentialType custom,omitempty,min=1"`
	Enabled        *bool    `json:"enabled" validate:"omitempty"`
}

type UpdateOAuthConnectionRequest struct {
	ProjectID      string   `uri:"project_id" validate:"required"`
	Provider       string   `uri:"provider" validate:"required,oneof=google"`
	CredentialType string   `json:"credential_type" validate:"required,oneof=default custom"`
	ClientID       string   `json:"client_id" validate:"required_if=CredentialType custom,notblank,max=512"`
	ClientSecret   string   `json:"client_secret" validate:"required_if=CredentialType custom,notblank,max=512"`
	Scopes         []string `json:"scopes" validate:"required_if=CredentialType custom,omitempty,min=1"`
	Enabled        *bool    `json:"enabled" validate:"omitempty"`
}

type DeleteOAuthConnectionRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	Provider  string `uri:"provider" validate:"required,oneof=google"`
}
