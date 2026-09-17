package models

import "github.com/roledio/roled/auth/internal/entities"

type VerifyEmailRequest struct {
	Token string `uri:"token" validate:"required"`
}

type VerifyEmailResult struct {
	Email    string
	Project  *entities.Project
	LoginURL *string
}
