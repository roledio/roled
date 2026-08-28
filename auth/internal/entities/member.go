package entities

import (
	"time"
)

type Member struct {
	ID        string     `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	AccountID string     `db:"account_id"`
	UserID    string     `db:"user_id"`
	IsAdmin   bool       `db:"is_admin"` // To indicate if the member is an admin of the account
}
