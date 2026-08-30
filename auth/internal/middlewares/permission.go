package middlewares

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func Permission(registry repositories.Registry) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()

		path := c.Path()
		if strings.HasSuffix(path, "/current") {
			// Skip permission check for routes that end with "/current" since it only requires a valid access token
			return c.Next()
		}

		// Retrieve access token from context
		localAccessToken := c.Locals(constants.CtxAccessToken)
		if localAccessToken == nil {
			return responseutil.SendError(c, errors.ErrInvalidAuthorizationToken)
		}

		// Fetch permissions associated with the access token
		var permissions []interfaces.PermissionResource
		permissionsRepo := registry.PermissionRepository()
		accessToken := localAccessToken.(*entities.AccessToken)
		if accessToken.UserID != nil {
			role, err := registry.RoleRepository().FindByUserID(ctx, *accessToken.UserID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find role by user ID", "error", err)
				return responseutil.SendError(c, errors.ErrSystemError.WithError(err))
			}
			if role != nil {
				// Get permissions by role if assigned, it's possible that a user doesn't have a role
				results, err := permissionsRepo.FindByRoleID(ctx, role.ID)
				if err != nil {
					log.WithContext(ctx).Errorw("Failed to find permissions by user ID", "error", err)
					return responseutil.SendError(c, errors.ErrSystemError.WithError(err))
				}
				permissions = results
			}
		} else {
			// Check client permissions
			results, err := permissionsRepo.FindByClientID(ctx, accessToken.ClientID)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to find permissions by client ID", "error", err)
				return responseutil.SendError(c, errors.ErrSystemError.WithError(err))
			}
			permissions = results
		}

		// Build a set of permission names for easy lookup
		permissionSet := make(map[string]any)
		for _, p := range permissions {
			name := p.ResourceCode + ":" + p.Code
			permissionSet[name] = p
		}

		// Store permission set in context for downstream handlers
		c.SetContext(context.WithValue(ctx, constants.CtxPermissions, permissionSet))

		// Get required permissions for the route
		routePermissions, ok := routePermissionsMap[c.Route().Name]
		if !ok || len(routePermissions) == 0 {
			// No permissions required for this route
			return c.Next()
		}

		// Check if any of the required permissions for this route are present
		for _, p := range routePermissions {
			if _, found := permissionSet[p]; found {
				// Permission granted
				return c.Next()
			}
		}

		// If we reach here, permission is denied
		return responseutil.SendError(c, errors.ErrInsufficientPermission)
	}
}

var routePermissionsMap = map[string][]string{
	// Account routes
	constants.RouteGetCurrentAccountDetails: {constants.PermissionReadAccounts},
	constants.RouteGetAccountDetails:        {constants.PermissionReadAccounts},
	constants.RouteGetAccounts:              {constants.PermissionReadAccounts},
	constants.RouteUpdateAccounts:           {constants.PermissionUpdateAccounts},
	constants.RouteDeleteAccounts:           {constants.PermissionDeleteAccounts},

	// Member routes
	constants.RouteGetMembers:       {constants.PermissionReadMembers},
	constants.RouteGetMemberDetails: {constants.PermissionReadMembers},
	constants.RouteCreateMembers:    {constants.PermissionCreateMembers},
	constants.RouteDeleteMembers:    {constants.PermissionDeleteMembers},

	// Project routes
	constants.RouteGetProjects:       {constants.PermissionReadProjects},
	constants.RouteGetProjectDetails: {constants.PermissionReadProjects},
	constants.RouteGetProjectByCode:  {constants.PermissionReadProjects},
	constants.RouteCreateProjects:    {constants.PermissionCreateProjects},
	constants.RouteUpdateProjects:    {constants.PermissionUpdateProjects},
	constants.RouteDeleteProjects:    {constants.PermissionDeleteProjects},

	// Project setting routes
	constants.RouteGetProjectSettings:    {constants.PermissionReadProjects},
	constants.RouteUpdateProjectSettings: {constants.PermissionUpdateProjects},

	// Project signup role routes
	constants.RouteUpdateProjectSignupRole: {constants.PermissionUpdateProjects},

	// Client routes
	constants.RouteGetProjectClients:       {constants.PermissionReadClients},
	constants.RouteGetProjectClientDetails: {constants.PermissionReadClients},
	constants.RouteGetProjectClientByCode:  {constants.PermissionReadClients},
	constants.RouteCreateProjectClients:    {constants.PermissionCreateClients},
	constants.RouteUpdateProjectClients:    {constants.PermissionUpdateClients},
	constants.RouteDeleteProjectClients:    {constants.PermissionDeleteClients},

	// Resource routes
	constants.RouteGetProjectResources:       {constants.PermissionReadResources},
	constants.RouteGetProjectResourceDetails: {constants.PermissionReadResources},
	constants.RouteCreateProjectResources:    {constants.PermissionCreateResources},
	constants.RouteUpdateProjectResources:    {constants.PermissionUpdateResources},
	constants.RouteDeleteProjectResources:    {constants.PermissionDeleteResources},

	// Permission routes
	constants.RouteGetProjectPermissions: {constants.PermissionReadPermissions},

	// Role routes
	constants.RouteGetProjectRoles:       {constants.PermissionReadRoles},
	constants.RouteGetProjectRoleDetails: {constants.PermissionReadRoles},
	constants.RouteCreateProjectRoles:    {constants.PermissionCreateRoles},
	constants.RouteUpdateProjectRoles:    {constants.PermissionUpdateRoles},
	constants.RouteDeleteProjectRoles:    {constants.PermissionDeleteRoles},

	// User routes
	constants.RouteGetProjectUsers:               {constants.PermissionReadUsers},
	constants.RouteGetProjectUserDetails:         {constants.PermissionReadUsers},
	constants.RouteGetProjectExternalUserDetails: {constants.PermissionReadUsers},
	constants.RouteCreateProjectUsers:            {constants.PermissionCreateUsers},
	constants.RouteUpdateProjectUsers:            {constants.PermissionUpdateUsers},
	constants.RouteDeleteProjectUsers:            {constants.PermissionDeleteUsers},
	constants.RouteResendVerificationEmail:       {constants.PermissionReadUsers},
	constants.RouteRequestPasswordReset:          {constants.PermissionReadUsers},
}
