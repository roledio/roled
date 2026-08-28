package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/utils/requestutil"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func (h *handler) getProjectPermissions(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetPermissionsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	permissions, total, err := h.permissionService.GetPermissions(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(permissions), total)
	return responseutil.SendSuccessWithPagination(c, permissions, pagination)
}
