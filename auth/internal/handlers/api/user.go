package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getProjectUsers(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetUsersRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	users, total, err := h.userService.GetUsers(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(users), total)
	return responseutil.SendSuccessWithPagination(c, users, pagination)
}

func (h *handler) getProjectUserDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetUserDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	user, err := h.userService.GetUserDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) getProjectExternalUserDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetExternalUserDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	user, err := h.userService.GetExternalUserDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) createProjectUser(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateUserRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	user, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) updateProjectUser(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateUserRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	user, err := h.userService.UpdateUser(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) deleteProjectUser(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteUserRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.userService.DeleteUser(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}

func (h *handler) getCurrentUserDetails(c fiber.Ctx) error {
	includePermissions, _ := strconv.ParseBool(c.Query("include_permissions", "false"))
	user, err := h.userService.GetCurrentUserDetails(c.Context(), includePermissions)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) updateCurrentUser(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateCurrentUserRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	user, err := h.userService.UpdateCurrentUser(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, user)
}

func (h *handler) resendVerificationEmail(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.ResendVerificationEmailRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.userService.ResendVerificationEmail(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}

func (h *handler) requestPasswordReset(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.RequestPasswordResetRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.userService.RequestPasswordReset(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
