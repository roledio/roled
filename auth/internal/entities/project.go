package entities

import (
	"time"
)

type Project struct {
	ID          string     `db:"id"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	AccountID   string     `db:"account_id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	LogoURL     *string    `db:"logo_url"`
	IsActive    bool       `db:"is_active"`
	IsSystem    bool       `db:"is_system"`
}
