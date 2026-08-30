package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getProjectClients(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetClientsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	clients, total, err := h.clientService.GetClients(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(clients), total)
	return responseutil.SendSuccessWithPagination(c, clients, pagination)
}

func (h *handler) getProjectClientDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetClientDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	client, err := h.clientService.GetClientDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, client)
}

func (h *handler) createProjectClient(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateClientRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	client, err := h.clientService.CreateClient(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, client)
}

func (h *handler) updateProjectClient(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateClientRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	client, err := h.clientService.UpdateClient(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, client)
}

func (h *handler) deleteProjectClient(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteClientRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.clientService.DeleteClient(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
