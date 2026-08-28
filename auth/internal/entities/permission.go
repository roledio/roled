package entities

import (
	"time"
)

type Permission struct {
	ID          string    `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	ResourceID  string    `db:"resource_id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsDefault   bool      `db:"is_default"` // whether this permission is included by default in new projects
}
