package models

import (
	"time"

	"github.com/roledio/roled/pkg/models"
)

type GetProjectRolesRequest struct {
	models.PageRequest
	ProjectID string `uri:"project_id" validate:"required"`
	Search    string `query:"search"`
}

type GetRoleDetailsRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	RoleID    string `uri:"role_id" validate:"required"`
}

type RoleDetails struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsDefaultSignup *bool     `json:"is_default_signup,omitempty"`
	PermissionIDs   []string  `json:"permission_ids,omitempty"`
}

type RoleDetailsAndPermissions struct {
	RoleDetails
	Permissions []RolePermission `json:"permissions"`
}

type RolePermission struct {
	ID             string `json:"id"`
	ResourceName   string `json:"resource_name"`
	PermissionName string `json:"permission_name"`
}

type CreateRoleRequest struct {
	ProjectID     string   `uri:"project_id" validate:"required"`
	Code          string   `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
	Name          string   `json:"name" validate:"required,notblank,max=50"`
	Description   string   `json:"description" validate:"omitempty,max=200"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	ProjectID     string   `uri:"project_id" validate:"required"`
	RoleID        string   `uri:"role_id" validate:"required"`
	Code          string   `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
	Name          string   `json:"name" validate:"required,notblank,max=50"`
	Description   string   `json:"description" validate:"omitempty,max=200"`
	PermissionIDs []string `json:"permission_ids"`
}

type DeleteRoleRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
	RoleID    string `uri:"role_id" validate:"required"`
}
