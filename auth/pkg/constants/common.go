package constants

import "github.com/roledio/roled/auth/pkg/types"

var (
	CtxRequestID types.ContextKey = "request_id"
	StrRequestID                  = "request_id"

	SystemStatusUp   = "up"
	SystemStatusDown = "down"
)
