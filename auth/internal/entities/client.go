package entities

import (
	"time"
)

type Client struct {
	ID              string     `db:"id"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	AccountID       string     `db:"account_id"`
	ProjectID       string     `db:"project_id"`
	Name            string     `db:"name"`
	Description     *string    `db:"description"`
	SecretEncrypted string     `db:"secret_encrypted"`
	IsActive        bool       `db:"is_active"`
	IsDefault       bool       `db:"is_default"`
}
