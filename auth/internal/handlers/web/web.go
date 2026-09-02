package web

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/authorize"
	"github.com/roledio/roled/auth/internal/services/infra"
	"github.com/roledio/roled/auth/internal/services/member"
	"github.com/roledio/roled/auth/internal/services/user"
)

type Dependencies struct {
	Registry         repositories.Registry
	Redis            infra.RedisService
	AuthorizeService authorize.AuthorizeService
	UserService      user.UserService
	MemberService    member.MemberService
}

type handler struct {
	app              *fiber.App
	defaultConfig    *configs.DefaultConfig
	repo             repositories.Registry
	redisService     infra.RedisService
	authorizeService authorize.AuthorizeService
	userService      user.UserService
	memberService    member.MemberService
}

func NewHandler(app *fiber.App, defaultConfig *configs.DefaultConfig, deps *Dependencies) *handler {
	return &handler{
		app:              app,
		defaultConfig:    defaultConfig,
		repo:             deps.Registry,
		redisService:     deps.Redis,
		authorizeService: deps.AuthorizeService,
		userService:      deps.UserService,
		memberService:    deps.MemberService,
	}
}

func (h *handler) SetupRoutes() {
	h.app.Get("/authorize", h.renderAuthorize)
	h.app.Post("/authorize", h.submitAuthorize)

	h.app.Get("/email/verify/:token", h.renderVerifyEmail)

	h.app.Get("/password/forgot", h.renderForgotPassword)
	h.app.Post("/password/forgot", h.submitForgotPassword)
	h.app.Get("/password/reset/:token", h.renderResetPassword)
	h.app.Post("/password/reset/:token", h.submitResetPassword)

	h.app.Get("/member/activate/:token", h.renderActivateMember)
	h.app.Post("/member/activate/:token", h.submitActivateMember)

	h.app.Get("/user/activate/:token", h.renderActivateProjectUser)
	h.app.Post("/user/activate/:token", h.submitActivateProjectUser)
}
