package constants

import "github.com/roledio/roled/auth/pkg/types"

const (
	RoledConsoleDefaultRoleCode = "default"

	CtxAccessToken types.ContextKey = "ctx_access_token"
	CtxPermissions types.ContextKey = "ctx_permissions"
	CtxAccount     types.ContextKey = "ctx_account"
)
