package models

import "github.com/roledio/roled/internal/entities"

type SendEmailRequest struct {
	To      []string
	CC      []string
	BCC     []string
	From    string
	Subject string
	Body    string
	IsHTML  bool
	Sender  string // Optional Sender header to set in the email
}

type VerifyEmailRequest struct {
	Token string `uri:"token" validate:"required"`
}

type VerifyEmailResult struct {
	Email    string
	Project  *entities.Project
	LoginURL *string
}
