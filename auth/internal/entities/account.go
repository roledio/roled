package entities

import (
	"time"
)

type Account struct {
	ID          string     `db:"id"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	IsActive    bool       `db:"is_active"`
	IsSystem    bool       `db:"is_system"`
}
