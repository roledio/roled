package entities

import (
	"time"
)

type OAuthConnection struct {
	ID                    string    `db:"id"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
	ProjectID             string    `db:"project_id"`
	Provider              string    `db:"provider"`
	CredentialType        string    `db:"credential_type"` // default, custom
	ClientID              *string   `db:"client_id"`
	ClientSecretEncrypted *string   `db:"client_secret_encrypted"`
	Scopes                *string   `db:"scopes"`
	Enabled               bool      `db:"enabled"`
}
