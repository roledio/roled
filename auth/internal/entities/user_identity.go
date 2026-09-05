package entities

import (
	"time"
)

type UserIdentity struct {
	ID             string     `db:"id"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
	UserID         string     `db:"user_id"`
	Provider       string     `db:"provider"`
	ProviderUserID string     `db:"provider_user_id"`
}
