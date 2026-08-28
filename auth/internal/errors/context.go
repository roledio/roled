package errors

import (
	"errors"
)

// Context-related errors for retrieving account and access token from context.
// These errors  are not expected to occur during normal operation if the context is properly set up.
// These errors will be handled by the responseutil and will be returned as system error.

var (
	ErrCtxAccountNotFound     = errors.New("account not found in request context")
	ErrCtxAccessTokenNotFound = errors.New("access token not found in request context")
)
