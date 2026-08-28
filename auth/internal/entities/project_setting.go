package entities

import (
	"time"
)

type ProjectSetting struct {
	ID                      string    `db:"id"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
	ProjectID               string    `db:"project_id"`
	IsSignupEnabled         bool      `db:"is_signup_enabled"`
	DefaultSignupRoleID     *string   `db:"default_signup_role_id"`
	IsSignupVerifyEmail     bool      `db:"is_signup_verify_email"`
	IsForgotPasswordEnabled bool      `db:"is_forgot_password_enabled"`
	IsAllowTempEmail        bool      `db:"is_allow_temp_email"`
}
