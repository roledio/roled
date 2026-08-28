package errors

import (
	"net/http"
)

type CustomError struct {
	Code         string
	Msg          string
	HttpCode     int
	Err          error
	DebugMessage string // Additional error message for debugging purpose when Err is not provided
}

func (e CustomError) Unwrap() error {
	return e.Err
}

func (e CustomError) Error() string {
	return e.Msg
}

func (e CustomError) WithError(err error) CustomError {
	if err != nil {
		e.Err = err
	}
	return e
}

func (e CustomError) WithDebugMessage(msg string) CustomError {
	e.DebugMessage = msg
	return e
}

var (
	ErrSystemError = CustomError{
		Code:     "system_error",
		Msg:      "An unexpected error occurred. Please try again later.",
		HttpCode: http.StatusInternalServerError,
	}
	ErrInvalidParams = CustomError{
		Code:     "invalid_params",
		Msg:      "Invalid or missing request parameters, headers, or body.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidAuthorizationToken = CustomError{
		Code:     "invalid_authorization_token",
		Msg:      "Unauthorized request. The authorization token is missing or invalid.",
		HttpCode: http.StatusUnauthorized,
	}
	ErrInsufficientPermission = CustomError{
		Code:     "insufficient_permission",
		Msg:      "Insufficient permissions to access this resource.",
		HttpCode: http.StatusForbidden,
	}
	ErrInvalidCSRFToken = CustomError{
		Code:     "invalid_csrf_token",
		Msg:      "Invalid security token. Please refresh the page and try again.",
		HttpCode: http.StatusForbidden,
	}
	ErrOperationNotAvailable = CustomError{
		Code:     "operation_not_available",
		Msg:      "The requested operation is not available.",
		HttpCode: http.StatusNotFound,
	}
	ErrFileSizeTooLarge = CustomError{
		Code:     "file_size_too_large",
		Msg:      "The uploaded file exceeds the maximum allowed file size.",
		HttpCode: http.StatusBadRequest,
	}
)
