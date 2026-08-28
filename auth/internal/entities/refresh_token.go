package entities

import (
	"time"
)

type RefreshToken struct {
	ID               string     `db:"id"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	AccountID        string     `db:"account_id"`
	ProjectID        string     `db:"project_id"`
	ClientID         string     `db:"client_id"`
	UserID           *string    `db:"user_id"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	Status           string     `db:"status"` // issued, used, expired, revoked
	ExpiresIn        *int       `db:"expires_in"`
	UsedAt           *time.Time `db:"used_at"`
	IssuedAt         *time.Time `db:"issued_at"`
	RevokedAt        *time.Time `db:"revoked_at"`
}
