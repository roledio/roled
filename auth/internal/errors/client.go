package errors

import (
	"net/http"

	"github.com/roledio/roled/pkg/errors"
)

var (
	ErrInvalidClientID = errors.CustomError{
		Code:     "invalid_client",
		Msg:      "The specified client ID is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrClientNotActive = errors.CustomError{
		Code:     "client_not_active",
		Msg:      "This client is currently inactive.",
		HttpCode: http.StatusBadRequest,
	}
	ErrClientNotFound = errors.CustomError{
		Code:     "client_not_found",
		Msg:      "The requested client could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrClientCodeAlreadyUsed = errors.CustomError{
		Code:     "client_code_already_used",
		Msg:      "A client with this code already exists in the project.",
		HttpCode: http.StatusConflict,
	}
	ErrDeleteDefaultClient = errors.CustomError{
		Code:     "delete_default_client_forbidden",
		Msg:      "The default client cannot be deleted.",
		HttpCode: http.StatusForbidden,
	}
)
