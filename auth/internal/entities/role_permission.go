package entities

import (
	"time"
)

type RolePermission struct {
	CreatedAt    time.Time `db:"created_at"`
	RoleID       string    `db:"role_id"`
	PermissionID string    `db:"permission_id"`
}
