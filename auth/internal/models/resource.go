package models

import (
	"time"

	"github.com/roledio/roled/pkg/models"
)

type GetResourcesRequest struct {
	models.PageRequest
	ProjectID string `uri:"project_id" validate:"required"`
	Search    string `query:"search"`
	IsDefault *bool  `query:"is_default"`
}

type ResourceDetails struct {
	ID          string       `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Description *string      `json:"description"`
	IsDefault   bool         `json:"is_default"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	IsDefault   bool    `json:"is_default"`
}

type GetResourceDetailsRequest struct {
	ProjectID  string `uri:"project_id" validate:"required"`
	ResourceID string `uri:"resource_id" validate:"required"`
}

type CreateResourceRequest struct {
	ProjectID   string  `uri:"project_id" validate:"required"`
	Name        string  `json:"name" validate:"required,notblank,max=50"`
	Code        string  `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
	Description *string `json:"description" validate:"omitempty,max=200"`
	Permissions []struct {
		Name        string  `json:"name" validate:"required,notblank,max=50"`
		Code        string  `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
		Description *string `json:"description" validate:"omitempty,max=200"`
	} `json:"permissions"`
}

type UpdateResourceRequest struct {
	ProjectID   string  `uri:"project_id" validate:"required"`
	ResourceID  string  `uri:"resource_id" validate:"required"`
	Name        string  `json:"name" validate:"required,notblank,max=50"`
	Code        string  `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
	Description *string `json:"description" validate:"omitempty,max=200"`
	Permissions []struct {
		Name        string  `json:"name" validate:"required,notblank,max=50"`
		Code        string  `json:"code" validate:"required,notblank,max=50,alphanum_dash_underscore"`
		Description *string `json:"description" validate:"omitempty,max=200"`
	} `json:"permissions" validate:"omitempty,dive"`
}

type DeleteResourceRequest struct {
	ProjectID  string `uri:"project_id" validate:"required"`
	ResourceID string `uri:"resource_id" validate:"required"`
}
