package errors

import (
	"net/http"
	"strings"

	"github.com/roledio/roled/auth/pkg/errors"
)

var (
	ErrSomePermissionsNotFound = func(missingIDs []string) errors.CustomError {
		return errors.CustomError{
			Code:     "some_permissions_not_found",
			Msg:      "The following permissions could not be found: " + strings.Join(missingIDs, ", "),
			HttpCode: http.StatusNotFound,
		}
	}
)
