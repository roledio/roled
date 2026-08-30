package errors

import (
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrInvalidRedirectURI = errors.CustomError{
		Code:     "invalid_redirect_uri",
		Msg:      "The specified redirect URI is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrRedirectURINotFound = errors.CustomError{
		Code:     "redirect_uri_not_found",
		Msg:      "The specified redirect URI could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrProjectNotActive = errors.CustomError{
		Code:     "project_not_active",
		Msg:      "This project is currently inactive.",
		HttpCode: http.StatusBadRequest,
	}
	ErrProjectNotFound = errors.CustomError{
		Code:     "project_not_found",
		Msg:      "The requested project could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrInvalidProjectCode = errors.CustomError{
		Code:     "invalid_project_code",
		Msg:      "The specified project code is invalid.",
		HttpCode: http.StatusNotFound,
	}
	ErrProjectCodeAlreadyUsed = errors.CustomError{
		Code:     "project_code_already_used",
		Msg:      "A project with this code already exists in the account.",
		HttpCode: http.StatusConflict,
	}
	ErrMoveTmpProjectLogo = errors.CustomError{
		Code:     "move_tmp_project_logo_failed",
		Msg:      "Unable to process the project logo URL. The file may not exist or has been moved. Please try again with a valid URL.",
		HttpCode: http.StatusInternalServerError,
	}
	ErrProjectNameRequiredForDeletion = errors.CustomError{
		Code:     "project_name_required_for_deletion",
		Msg:      "The project name is required and must match the existing project name to confirm deletion.",
		HttpCode: http.StatusBadRequest,
	}
	ErrModifySystemProject = errors.CustomError{
		Code:     "modify_system_project_forbidden",
		Msg:      "System projects cannot be modified or deleted.",
		HttpCode: http.StatusForbidden,
	}
	ErrProjectSettingsNotFound = errors.CustomError{
		Code:     "project_settings_not_found",
		Msg:      "Project settings could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrDefaultSignupRoleRequired = errors.CustomError{
		Code:     "default_signup_role_required",
		Msg:      "A default sign-up role is required when sign-up is enabled.",
		HttpCode: http.StatusBadRequest,
	}
	ErrSignupMustBeEnabled = errors.CustomError{
		Code:     "signup_must_be_enabled",
		Msg:      "Sign-up must be enabled before assigning a default sign-up role.",
		HttpCode: http.StatusBadRequest,
	}
)
