package errors

import (
	"fmt"
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrOAuthConnectionNotFound = errors.CustomError{
		Code:     "oauth_connection_not_found",
		Msg:      "The requested OAuth connection could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrOAuthConnectionAlreadyExists = errors.CustomError{
		Code:     "oauth_connection_already_exists",
		Msg:      "An OAuth connection for this provider already exists.",
		HttpCode: http.StatusConflict,
	}
	ErrProviderOAuthConnectionNotFound = func(provider string) error {
		return errors.CustomError{
			Code:     "provider_oauth_connection_not_found",
			Msg:      fmt.Sprintf("The OAuth connection for provider \"%s\" could not be found for this project.", provider),
			HttpCode: http.StatusNotFound,
		}
	}
	ErrProviderOAuthConnectionDisabled = func(provider string) error {
		return errors.CustomError{
			Code:     "provider_oauth_connection_disabled",
			Msg:      fmt.Sprintf("The OAuth connection for provider \"%s\" is disabled for this project.", provider),
			HttpCode: http.StatusBadRequest,
		}
	}
)
