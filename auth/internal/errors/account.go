package errors

import (
	"net/http"

	"github.com/roledio/roled/pkg/errors"
)

var (
	ErrAccountNotFound = errors.CustomError{
		Code:     "account_not_found",
		Msg:      "The requested account could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrAccountNotActive = errors.CustomError{
		Code:     "account_not_active",
		Msg:      "This account is currently inactive.",
		HttpCode: http.StatusBadRequest,
	}
	ErrModifySystemAccount = errors.CustomError{
		Code:     "modify_system_account_forbidden",
		Msg:      "System accounts cannot be modified or deleted.",
		HttpCode: http.StatusForbidden,
	}
	ErrNonUserDeleteAccount = errors.CustomError{
		Code:     "account_deletion_by_non_user_forbidden",
		Msg:      "Account deletion must be initiated by an authenticated user.",
		HttpCode: http.StatusUnauthorized,
	}
	ErrNonAdminDeleteAccount = errors.CustomError{
		Code:     "account_deletion_by_non_admin_forbidden",
		Msg:      "Only account administrators can delete an account.",
		HttpCode: http.StatusForbidden,
	}
)
