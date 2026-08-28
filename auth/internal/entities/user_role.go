package entities

import (
	"time"
)

type UserRole struct {
	CreatedAt time.Time `db:"created_at"`
	UserID    string    `db:"user_id"`
	RoleID    string    `db:"role_id"`
}
