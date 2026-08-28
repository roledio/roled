package entities

import (
	"time"
)

type ClientPermission struct {
	CreatedAt    time.Time `db:"created_at"`
	ClientID     string    `db:"client_id"`
	PermissionID string    `db:"permission_id"`
}
