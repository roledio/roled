package models

type GetProjectSettingsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
}

type ProjectSettings struct {
	IsSignupEnabled         bool    `json:"is_signup_enabled"`
	DefaultSignupRoleID     *string `json:"default_signup_role_id"`
	IsSignupVerifyEmail     bool    `json:"is_signup_verify_email"`
	IsForgotPasswordEnabled bool    `json:"is_forgot_password_enabled"`
	IsAllowTempEmail        bool    `json:"is_allow_temp_email"`
}

type UpdateProjectSettingsRequest struct {
	ProjectID               string  `uri:"project_id" validate:"required"`
	IsSignupEnabled         *bool   `json:"is_signup_enabled" validate:"required"`
	DefaultSignupRoleID     *string `json:"default_signup_role_id" validate:"omitempty"`
	IsSignupVerifyEmail     *bool   `json:"is_signup_verify_email" validate:"required"`
	IsForgotPasswordEnabled *bool   `json:"is_forgot_password_enabled" validate:"required"`
	IsAllowTempEmail        *bool   `json:"is_allow_temp_email" validate:"required"`
}

type UpdateProjectSignupRoleRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	RoleID    string `json:"role_id" validate:"required,notblank"`
}

type UpdateProjectSignupRoleResponse struct {
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
}
