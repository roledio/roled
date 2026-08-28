package errors

import (
	"net/http"

	"github.com/roledio/roled/pkg/errors"
)

var (
	ErrInvalidGrantType = errors.CustomError{
		Code:     "invalid_grant_type",
		Msg:      "The specified grant type is invalid or unsupported.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidRefreshToken = errors.CustomError{
		Code:     "invalid_refresh_token",
		Msg:      "The provided refresh token is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidAuthorizationCode = errors.CustomError{
		Code:     "invalid_authorization_code",
		Msg:      "The provided authorization code is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidCodeVerifier = errors.CustomError{
		Code:     "invalid_code_verifier",
		Msg:      "The provided code verifier is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrAuthCodeAlreadyUsed = errors.CustomError{
		Code:     "auth_code_already_used",
		Msg:      "This authorization code has already been used.",
		HttpCode: http.StatusBadRequest,
	}
	ErrAuthCodeExpired = errors.CustomError{
		Code:     "auth_code_expired",
		Msg:      "This authorization code has expired.",
		HttpCode: http.StatusBadRequest,
	}
	ErrRefreshTokenExpired = errors.CustomError{
		Code:     "refresh_token_expired",
		Msg:      "This refresh token has expired. Please sign in again.",
		HttpCode: http.StatusBadRequest,
	}
	ErrRefreshTokenAlreadyUsed = errors.CustomError{
		Code:     "refresh_token_already_used",
		Msg:      "This refresh token has already been used.",
		HttpCode: http.StatusBadRequest,
	}
	ErrDifferentAuthorizeRedirectURI = errors.CustomError{
		Code:     "different_authorize_redirect_uri",
		Msg:      "The redirect URI does not match the URI used in the initial authorization request.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidClientCredentials = errors.CustomError{
		Code:     "invalid_client_credentials",
		Msg:      "Invalid client credentials provided.",
		HttpCode: http.StatusUnauthorized,
	}
)
