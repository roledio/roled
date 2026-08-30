package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getProjectResources(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetResourcesRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	resources, total, err := h.resourceService.GetResources(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(resources), total)
	return responseutil.SendSuccessWithPagination(c, resources, pagination)
}

func (h *handler) getResourceDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetResourceDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	result, err := h.resourceService.GetResourceDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, result)
}

func (h *handler) createProjectResource(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateResourceRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	resource, err := h.resourceService.CreateResource(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, resource)
}

func (h *handler) updateProjectResource(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateResourceRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	resource, err := h.resourceService.UpdateResource(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, resource)
}

func (h *handler) deleteProjectResource(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteResourceRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.resourceService.DeleteResource(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
