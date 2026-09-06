package errors

import (
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrOAuthConnectionNotFound = errors.CustomError{
		Code:     "oauth_connection_not_found",
		Msg:      "The requested OAuth connection could not be found.",
		HttpCode: http.StatusNotFound,
	}
)
