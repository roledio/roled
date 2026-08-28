package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/middlewares"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/accesstoken"
	"github.com/roledio/roled/internal/services/account"
	"github.com/roledio/roled/internal/services/client"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/services/member"
	"github.com/roledio/roled/internal/services/permission"
	"github.com/roledio/roled/internal/services/project"
	"github.com/roledio/roled/internal/services/resource"
	"github.com/roledio/roled/internal/services/role"
	"github.com/roledio/roled/internal/services/upload"
	"github.com/roledio/roled/internal/services/user"
)

type Dependencies struct {
	Registry          repositories.Registry
	Redis             infra.RedisService
	ProjectService    project.ProjectService
	TokenService      accesstoken.AccessTokenService
	AccountService    account.AccountService
	MemberService     member.MemberService
	UploadService     upload.UploadService
	ClientService     client.ClientService
	ResourceService   resource.ResourceService
	RoleService       role.RoleService
	UserService       user.UserService
	PermissionService permission.PermissionService
}

type handler struct {
	app               *fiber.App
	defaultConfig     *configs.DefaultConfig
	repo              repositories.Registry
	redisService      infra.RedisService
	projectService    project.ProjectService
	tokenService      accesstoken.AccessTokenService
	accountService    account.AccountService
	memberService     member.MemberService
	uploadService     upload.UploadService
	clientService     client.ClientService
	resourceService   resource.ResourceService
	roleService       role.RoleService
	userService       user.UserService
	permissionService permission.PermissionService
}

func NewHandler(app *fiber.App, defaultConfig *configs.DefaultConfig, deps *Dependencies) *handler {
	return &handler{
		app:               app,
		defaultConfig:     defaultConfig,
		repo:              deps.Registry,
		redisService:      deps.Redis,
		projectService:    deps.ProjectService,
		tokenService:      deps.TokenService,
		accountService:    deps.AccountService,
		memberService:     deps.MemberService,
		uploadService:     deps.UploadService,
		clientService:     deps.ClientService,
		resourceService:   deps.ResourceService,
		roleService:       deps.RoleService,
		userService:       deps.UserService,
		permissionService: deps.PermissionService,
	}
}

func (h *handler) SetupRoutes() {
	h.app.Get("/system/ping", h.ping)
	h.app.Get("/system/health", h.getSystemHealth)
	h.app.Get("/system/info", h.getSystemInfo)
	h.app.Get("/system/console/config", h.getConsoleConfig)

	h.app.Post("/api/v1/tokens", h.exchangeToken)
	h.app.Post("/api/v1/tokens/current/revoke", h.revokeCurrentToken)
	h.protectedGet("/api/v1/tokens/current", constants.RouteGetCurrentTokenDetails, h.getCurrentAccessToken)

	h.protectedGet("/api/v1/accounts/current", constants.RouteGetCurrentAccountDetails, h.getCurrentAccountDetails)

	h.protectedGet("/api/v1/users/current", constants.RouteGetCurrentUserDetails, h.getCurrentUserDetails)
	h.protectedPut("/api/v1/users/current", constants.RouteUpdateCurrentUser, h.updateCurrentUser)

	h.protectedPost("/api/v1/uploads", constants.RouteUploadFiles, h.uploadFile)

	h.protectedGet("/api/v1/accounts/:account_id", constants.RouteGetAccountDetails, h.getAccountDetails)
	h.protectedGet("/api/v1/accounts", constants.RouteGetAccounts, h.getAccounts)
	h.protectedPut("/api/v1/accounts/:account_id", constants.RouteUpdateAccounts, h.updateAccount)
	h.protectedPost("/api/v1/accounts/:account_id/delete", constants.RouteDeleteAccounts, h.deleteAccount)

	h.protectedGet("/api/v1/members", constants.RouteGetMembers, h.getMembers)
	h.protectedGet("/api/v1/members/:member_id", constants.RouteGetMemberDetails, h.getMemberDetails)
	h.protectedPost("/api/v1/members", constants.RouteCreateMembers, h.createMember)
	h.protectedPatch("/api/v1/members/:member_id", constants.RouteUpdateMembers, h.updateMember)
	h.protectedDelete("/api/v1/members/:member_id", constants.RouteDeleteMembers, h.deleteMember)

	h.protectedGet("/api/v1/projects", constants.RouteGetProjects, h.getProjects)
	h.protectedGet("/api/v1/projects/:project_id", constants.RouteGetProjectDetails, h.getProjectDetails)
	h.protectedPost("/api/v1/projects", constants.RouteCreateProjects, h.createProject)
	h.protectedPut("/api/v1/projects/:project_id", constants.RouteUpdateProjects, h.updateProject)
	h.protectedPost("/api/v1/projects/:project_id/delete", constants.RouteDeleteProjects, h.deleteProject)

	h.protectedGet("/api/v1/projects/:project_id/settings", constants.RouteGetProjectSettings, h.getProjectSettings)
	h.protectedPut("/api/v1/projects/:project_id/settings", constants.RouteUpdateProjectSettings, h.updateProjectSettings)

	h.protectedPatch("/api/v1/projects/:project_id/signup-role", constants.RouteUpdateProjectSignupRole, h.updateProjectSignupRole)

	h.protectedGet("/api/v1/projects/:project_id/clients", constants.RouteGetProjectClients, h.getProjectClients)
	h.protectedGet("/api/v1/projects/:project_id/clients/:client_id", constants.RouteGetProjectClientDetails, h.getProjectClientDetails)
	h.protectedPost("/api/v1/projects/:project_id/clients", constants.RouteCreateProjectClients, h.createProjectClient)
	h.protectedPut("/api/v1/projects/:project_id/clients/:client_id", constants.RouteUpdateProjectClients, h.updateProjectClient)
	h.protectedDelete("/api/v1/projects/:project_id/clients/:client_id", constants.RouteDeleteProjectClients, h.deleteProjectClient)

	h.protectedGet("/api/v1/projects/:project_id/resources", constants.RouteGetProjectResources, h.getProjectResources)
	h.protectedGet("/api/v1/projects/:project_id/resources/:resource_id", constants.RouteGetProjectResourceDetails, h.getResourceDetails)
	h.protectedPost("/api/v1/projects/:project_id/resources", constants.RouteCreateProjectResources, h.createProjectResource)
	h.protectedPut("/api/v1/projects/:project_id/resources/:resource_id", constants.RouteUpdateProjectResources, h.updateProjectResource)
	h.protectedDelete("/api/v1/projects/:project_id/resources/:resource_id", constants.RouteDeleteProjectResources, h.deleteProjectResource)

	h.protectedGet("/api/v1/projects/:project_id/permissions", constants.RouteGetProjectPermissions, h.getProjectPermissions)

	h.protectedGet("/api/v1/projects/:project_id/roles", constants.RouteGetProjectRoles, h.getProjectRoles)
	h.protectedGet("/api/v1/projects/:project_id/roles/:role_id", constants.RouteGetProjectRoleDetails, h.getProjectRoleDetails)
	h.protectedPost("/api/v1/projects/:project_id/roles", constants.RouteCreateProjectRoles, h.createProjectRole)
	h.protectedPut("/api/v1/projects/:project_id/roles/:role_id", constants.RouteUpdateProjectRoles, h.updateProjectRole)
	h.protectedDelete("/api/v1/projects/:project_id/roles/:role_id", constants.RouteDeleteProjectRoles, h.deleteProjectRole)

	h.protectedGet("/api/v1/projects/:project_id/users", constants.RouteGetProjectUsers, h.getProjectUsers)
	h.protectedGet("/api/v1/projects/:project_id/users/:user_id", constants.RouteGetProjectUserDetails, h.getProjectUserDetails)
	h.protectedGet("/api/v1/projects/:project_id/users/external/:external_user_id", constants.RouteGetProjectExternalUserDetails, h.getProjectExternalUserDetails)
	h.protectedPost("/api/v1/projects/:project_id/users", constants.RouteCreateProjectUsers, h.createProjectUser)
	h.protectedPut("/api/v1/projects/:project_id/users/:user_id", constants.RouteUpdateProjectUsers, h.updateProjectUser)
	h.protectedDelete("/api/v1/projects/:project_id/users/:user_id", constants.RouteDeleteProjectUsers, h.deleteProjectUser)
	h.protectedPost("/api/v1/projects/:project_id/users/:user_id/verification-email", constants.RouteResendVerificationEmail, h.resendVerificationEmail)
	h.protectedPost("/api/v1/projects/:project_id/users/:user_id/password-reset", constants.RouteRequestPasswordReset, h.requestPasswordReset)
}

// protectedGet registers a protected GET route with JWT and Permission middlewares
func (h *handler) protectedGet(path string, name string, pathHandler fiber.Handler) {
	h.app.Get(path,
		middlewares.JWT(h.defaultConfig, h.repo),
		middlewares.Permission(h.repo),
		pathHandler).
		Name(name)
}

// protectedPost registers a protected POST route with JWT and Permission middlewares
func (h *handler) protectedPost(path string, name string, pathHandler fiber.Handler) {
	h.app.Post(path,
		middlewares.JWT(h.defaultConfig, h.repo),
		middlewares.Permission(h.repo),
		pathHandler).
		Name(name)
}

// protectedPut registers a protected PUT route with JWT and Permission middlewares
func (h *handler) protectedPut(path string, name string, pathHandler fiber.Handler) {
	h.app.Put(path,
		middlewares.JWT(h.defaultConfig, h.repo),
		middlewares.Permission(h.repo),
		pathHandler).
		Name(name)
}

// protectedDelete registers a protected DELETE route with JWT and Permission middlewares
func (h *handler) protectedDelete(path string, name string, pathHandler fiber.Handler) {
	h.app.Delete(path,
		middlewares.JWT(h.defaultConfig, h.repo),
		middlewares.Permission(h.repo),
		pathHandler).
		Name(name)
}

// protectedPatch registers a protected PATCH route with JWT and Permission middlewares
func (h *handler) protectedPatch(path string, name string, pathHandler fiber.Handler) {
	h.app.Patch(path,
		middlewares.JWT(h.defaultConfig, h.repo),
		middlewares.Permission(h.repo),
		pathHandler).
		Name(name)
}
