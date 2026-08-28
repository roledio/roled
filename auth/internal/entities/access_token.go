package entities

import (
	"time"
)

type AccessToken struct {
	ID             string     `db:"id"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
	AccountID      string     `db:"account_id"`
	ProjectID      string     `db:"project_id"`
	ClientID       string     `db:"client_id"`
	UserID         *string    `db:"user_id"`
	RefreshTokenID *string    `db:"refresh_token_id"`
	AuthCodeID     *string    `db:"auth_code_id"`
	GrantType      string     `db:"grant_type"`
	Status         string     `db:"status"` // issued, expired, revoked
	ExpiresIn      *int       `db:"expires_in"`
	IssuedAt       *time.Time `db:"issued_at"`
	RevokedAt      *time.Time `db:"revoked_at"`
}
