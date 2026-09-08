package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getOAuthConnections(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetOAuthConnectionsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	connections, err := h.oAuthConnectionService.GetOAuthConnections(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, connections)
}

func (h *handler) getOAuthConnectionDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetOAuthConnectionRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	connection, err := h.oAuthConnectionService.GetOAuthConnectionDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, connection)
}

func (h *handler) createOAuthConnection(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.CreateOAuthConnectionRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	connection, err := h.oAuthConnectionService.CreateOAuthConnection(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, connection)
}

func (h *handler) updateOAuthConnection(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateOAuthConnectionRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	connection, err := h.oAuthConnectionService.UpdateOAuthConnection(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, connection)
}

func (h *handler) deleteOAuthConnection(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteOAuthConnectionRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.oAuthConnectionService.DeleteOAuthConnection(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
