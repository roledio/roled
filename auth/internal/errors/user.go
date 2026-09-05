package errors

import (
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrUserNotFound = errors.CustomError{
		Code:     "user_not_found",
		Msg:      "The requested user could not be found.",
		HttpCode: http.StatusNotFound,
	}
	ErrInvalidPassword = errors.CustomError{
		Code:     "invalid_password",
		Msg:      "The password provided is incorrect.",
		HttpCode: http.StatusUnauthorized,
	}
	ErrInvalidUserCredentials = errors.CustomError{
		Code:     "invalid_user_credentials",
		Msg:      "Invalid email or password.",
		HttpCode: http.StatusNotFound,
	}
	ErrUnableToProcessSignup = errors.CustomError{
		Code:     "unable_to_process_signup",
		Msg:      "Unable to complete sign-up. Please verify the submitted information and try again.",
		HttpCode: http.StatusUnprocessableEntity,
	}
	ErrUserEmailAlreadyUsed = errors.CustomError{
		Code:     "user_email_already_used",
		Msg:      "A user with this email address already exists.",
		HttpCode: http.StatusConflict,
	}
	ErrDisposableEmail = errors.CustomError{
		Code:     "email_not_allowed",
		Msg:      "This email domain is not allowed for registration.",
		HttpCode: http.StatusBadRequest,
	}
	ErrUserNotVerified = errors.CustomError{
		Code:     "user_not_verified",
		Msg:      "The email address is not verified. Email verification is required to proceed.",
		HttpCode: http.StatusForbidden,
	}
	ErrUserNotActive = errors.CustomError{
		Code:     "user_not_active",
		Msg:      "This user is currently inactive.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidVerifyEmailToken = errors.CustomError{
		Code:     "invalid_verify_email_token",
		Msg:      "The email verification link is invalid or has expired. Please request a new verification email.",
		HttpCode: http.StatusNotFound,
	}
	ErrInvalidResetPasswordToken = errors.CustomError{
		Code:     "invalid_reset_password_token",
		Msg:      "The password reset link is invalid or has expired. Please request a new password reset link.",
		HttpCode: http.StatusNotFound,
	}
	ErrUserExternalIDAlreadyUsed = errors.CustomError{
		Code:     "user_external_id_already_used",
		Msg:      "A user with this external ID already exists in the project.",
		HttpCode: http.StatusConflict,
	}
	ErrMoveTmpUserAvatar = errors.CustomError{
		Code:     "move_tmp_user_avatar_failed",
		Msg:      "Unable to process the avatar URL. The file may not exist or has been moved. Please try again with a valid URL.",
		HttpCode: http.StatusInternalServerError,
	}
	ErrUserHasNoEmail = errors.CustomError{
		Code:     "user_has_no_email",
		Msg:      "The user does not have an email address.",
		HttpCode: http.StatusBadRequest,
	}
	ErrUserAlreadyVerified = errors.CustomError{
		Code:     "user_already_verified",
		Msg:      "This user's email address is already verified.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidActivationToken = errors.CustomError{
		Code:     "invalid_activation_token",
		Msg:      "The activation link is invalid or has expired. Please request a new invitation.",
		HttpCode: http.StatusNotFound,
	}
	ErrUserAlreadyActive = errors.CustomError{
		Code:     "user_already_active",
		Msg:      "This user is already activated.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidGoogleState = errors.CustomError{
		Code:     "invalid_google_state",
		Msg:      "The Google OAuth state is invalid or has expired.",
		HttpCode: http.StatusBadRequest,
	}
	ErrGoogleTokenExchangeFailed = errors.CustomError{
		Code:     "google_token_exchange_failed",
		Msg:      "Failed to exchange Google authorization code.",
		HttpCode: http.StatusInternalServerError,
	}
	ErrGoogleIDTokenMissing = errors.CustomError{
		Code:     "google_id_token_missing",
		Msg:      "Google ID token is missing from the response.",
		HttpCode: http.StatusInternalServerError,
	}
	ErrGoogleIDTokenInvalid = errors.CustomError{
		Code:     "google_id_token_invalid",
		Msg:      "Google ID token is invalid.",
		HttpCode: http.StatusInternalServerError,
	}
)
