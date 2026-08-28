package user

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

// GetCurrentUserDetails retrieves the details of the currently authenticated user based on the provided request.
// The difference between this method and GetUserDetails is that this method does not perform project validation,
// as it assumes the user is already authenticated and authorized to access their own details.
//
// The GetUserDetails method requires project validation to ensure the user has access to the specified project based on
// the account the user belongs to, while GetCurrentUserDetails directly retrieves the user details without additional checks.
func (s *userService) GetCurrentUserDetails(ctx context.Context, includePermissions bool) (*models.UserDetails, error) {
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		return nil, errors.ErrCtxAccessTokenNotFound
	}
	if accessToken.UserID == nil {
		err := fmt.Errorf("non-user access token is not supported for getting current user details")
		log.WithContext(ctx).Errorw("Failed to get current user details", "error", err)
		return nil, pkgerrors.ErrOperationNotAvailable.WithError(err)
	}
	return s.processGetUserDetails(ctx, &models.GetUserDetailsRequest{
		ProjectID:          accessToken.ProjectID,
		UserID:             *accessToken.UserID,
		IncludePermissions: includePermissions,
	})
}
