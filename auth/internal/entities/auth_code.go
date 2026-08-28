package entities

import (
	"time"
)

type AuthCode struct {
	ID                  string     `db:"id"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	AccountID           string     `db:"account_id"`
	ProjectID           string     `db:"project_id"`
	ClientID            string     `db:"client_id"`
	UserID              *string    `db:"user_id"`
	CodeHash            string     `db:"code_hash"`
	CodeChallenge       string     `db:"code_challenge"`
	CodeChallengeMethod string     `db:"code_challenge_method"`
	RedirectURI         string     `db:"redirect_uri"`
	State               *string    `db:"state"`
	ExpiresAt           time.Time  `db:"expires_at"`
	UsedAt              *time.Time `db:"used_at"`
}
