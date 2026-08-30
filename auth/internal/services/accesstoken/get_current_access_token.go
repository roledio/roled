package accesstoken

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *accessTokenService) GetCurrentAccessToken(ctx context.Context) (*models.AccessTokenDetails, error) {
	// Get current access token from context, should not be nil here
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		return nil, errors.ErrCtxAccessTokenNotFound
	}

	result, err := s.registry.AccessTokenRepository().FindByIDJoin(ctx, accessToken.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find access token by ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	res := models.AccessTokenDetails{
		ID:        result.ID,
		IssuedAt:  result.IssuedAt,
		ExpiresAt: result.IssuedAt.Add(time.Duration(result.ExpiresIn) * time.Second),
		Project: models.AccessTokenProject{
			ID:          result.ProjectID,
			Name:        result.ProjectName,
			Description: result.ProjectDescription,
			LogoURL:     result.ProjectLogoURL,
		},
		Client: models.AccessTokenClient{
			ID:   result.ClientID,
			Name: result.ClientName,
		},
	}

	if result.UserID != nil {
		res.User = &models.AccessTokenUser{
			ID:             *result.UserID,
			DisplayName:    *result.UserDisplayName,
			Email:          result.UserEmail,
			ExternalUserID: result.UserExternalUserID,
			AvatarURL:      result.UserAvatarURL,
		}
	}

	if result.RoleID != nil {
		res.Role = &models.AccessTokenRole{
			ID:          *result.RoleID,
			Code:        *result.RoleCode,
			Name:        *result.RoleName,
			Description: *result.RoleDescription,
		}
	}

	var permissions []interfaces.PermissionResource
	permissionsRepo := s.registry.PermissionRepository()
	if accessToken.UserID != nil && result.RoleID != nil {
		// Check user role permissions
		results, err := permissionsRepo.FindByRoleID(ctx, *result.RoleID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find permissions by user ID", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		permissions = results
	} else {
		// Check client permissions
		results, err := permissionsRepo.FindByClientID(ctx, accessToken.ClientID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find permissions by client ID", "error", err)
			return nil, pkgerrors.ErrSystemError.WithError(err)
		}
		permissions = results
	}

	for _, p := range permissions {
		res.Permissions = append(res.Permissions, p.ResourceCode+":"+p.Code)
	}

	return &res, nil
}
