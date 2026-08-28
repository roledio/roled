package errors

import (
	"net/http"

	"github.com/roledio/roled/pkg/errors"
)

var (
	ErrRoleNotFound = errors.CustomError{
		Code:     "role_not_found",
		Msg:      "The requested role could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrRoleCodeAlreadyUsed = errors.CustomError{
		Code:     "role_code_already_used",
		Msg:      "A role with this code already exists in the project.",
		HttpCode: http.StatusConflict,
	}
)
