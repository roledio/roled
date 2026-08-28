package errors

import (
	"net/http"

	"github.com/roledio/roled/pkg/errors"
)

var (
	ErrMemberNotFound = errors.CustomError{
		Code:     "member_not_found",
		Msg:      "The requested account member could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrInvalidActivateMemberToken = errors.CustomError{
		Code: "invalid_activate_member_token",
		Msg: `The activation link is invalid or has expired. Please contact an 
		account administrator for assistance.`,
		HttpCode: http.StatusNotFound,
	}
	ErrNonAdminDeleteMember = errors.CustomError{
		Code:     "delete_member_by_non_admin_forbidden",
		Msg:      "Only account administrators can remove members.",
		HttpCode: http.StatusForbidden,
	}
	ErrCannotDeleteSelf = errors.CustomError{
		Code:     "delete_member_self_forbidden",
		Msg:      "Cannot remove the currently authenticated member.",
		HttpCode: http.StatusBadRequest,
	}
	ErrCannotDeleteLastAdmin = errors.CustomError{
		Code:     "delete_last_admin_member_forbidden",
		Msg:      "Cannot remove the last remaining administrator from the account.",
		HttpCode: http.StatusBadRequest,
	}
	ErrNonAdminUpdateMember = errors.CustomError{
		Code:     "update_member_by_non_admin_forbidden",
		Msg:      "Only account administrators can update members.",
		HttpCode: http.StatusForbidden,
	}
	ErrCannotUpdateSelf = errors.CustomError{
		Code:     "update_member_self_forbidden",
		Msg:      "Cannot change the admin status of the currently authenticated member.",
		HttpCode: http.StatusBadRequest,
	}
	ErrCannotDemoteLastAdmin = errors.CustomError{
		Code:     "demote_last_admin_member_forbidden",
		Msg:      "Cannot remove admin privileges from the last remaining administrator of the account.",
		HttpCode: http.StatusBadRequest,
	}
)
