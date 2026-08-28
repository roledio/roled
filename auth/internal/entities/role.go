package entities

import (
	"time"
)

type Role struct {
	ID          string     `db:"id"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	AccountID   string     `db:"account_id"`
	ProjectID   string     `db:"project_id"`
	Code        string     `db:"code"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
}
