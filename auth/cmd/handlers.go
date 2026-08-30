package main

import (
	"github.com/roledio/roled/auth/internal/handlers/api"
	"github.com/roledio/roled/auth/internal/handlers/web"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
)

func newApiHandlerDeps(registry repositories.Registry, redis infra.RedisService, services *Services) *api.Dependencies {
	return &api.Dependencies{
		Registry:          registry,
		Redis:             redis,
		ProjectService:    services.ProjectService,
		TokenService:      services.TokenService,
		AccountService:    services.AccountService,
		MemberService:     services.MemberService,
		UploadService:     services.UploadService,
		ClientService:     services.ClientService,
		ResourceService:   services.ResourceService,
		RoleService:       services.RoleService,
		UserService:       services.UserService,
		PermissionService: services.PermissionService,
	}
}

func newWebHandlerDeps(registry repositories.Registry, redis infra.RedisService, services *Services) *web.Dependencies {
	return &web.Dependencies{
		Registry:         registry,
		Redis:            redis,
		AuthorizeService: services.AuthorizeService,
		UserService:      services.UserService,
		MemberService:    services.MemberService,
	}
}
