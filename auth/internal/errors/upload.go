package errors

import (
	"net/http"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrInvalidUploadType = errors.CustomError{
		Code:     "invalid_upload_type",
		Msg:      "The specified upload type is invalid.",
		HttpCode: http.StatusBadRequest,
	}
	ErrInvalidImageType = errors.CustomError{
		Code:     "invalid_image_type",
		Msg:      "The uploaded image file type is invalid or unsupported.",
		HttpCode: http.StatusBadRequest,
	}
)
