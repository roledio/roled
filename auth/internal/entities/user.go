package entities

import (
	"time"
)

type User struct {
	ID              string     `db:"id"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	AccountID       string     `db:"account_id"`
	ProjectID       string     `db:"project_id"`
	Email           *string    `db:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	PasswordHash    *string    `db:"password_hash"`
	ExternalUserID  *string    `db:"external_user_id"`
	DisplayName     string     `db:"display_name"`
	AvatarURL       *string    `db:"avatar_url"`
	IsActive        bool       `db:"is_active"`
}
