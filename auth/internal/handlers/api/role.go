package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/utils/requestutil"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func (h *handler) getProjectRoles(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetProjectRolesRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	roles, total, err := h.roleService.GetRoles(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(roles), total)
	return responseutil.SendSuccessWithPagination(c, roles, pagination)
}

func (h *handler) getProjectRoleDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetRoleDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	role, err := h.roleService.GetRoleDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, role)
}

func (h *handler) createProjectRole(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateRoleRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	role, err := h.roleService.CreateRole(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, role)
}

func (h *handler) updateProjectRole(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateRoleRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	role, err := h.roleService.UpdateRole(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, role)
}

func (h *handler) deleteProjectRole(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteRoleRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.roleService.DeleteRole(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
