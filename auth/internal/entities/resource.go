package entities

import (
	"time"
)

type Resource struct {
	ID          string    `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	AccountID   string    `db:"account_id"`
	ProjectID   string    `db:"project_id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	IsDefault   bool      `db:"is_default"` // whether this resource is included by default in new projects
}
