package errors

import (
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrResourceNotFound = errors.CustomError{
		Code:     "resource_not_found",
		Msg:      "The requested resource could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrResourceCodeAlreadyUsed = errors.CustomError{
		Code:     "resource_code_already_used",
		Msg:      "A resource with this code already exists in the project.",
		HttpCode: http.StatusConflict,
	}
	ErrModifyDefaultResource = errors.CustomError{
		Code:     "modify_default_resource_forbidden",
		Msg:      "Default resources cannot be modified or deleted.",
		HttpCode: http.StatusForbidden,
	}
)
