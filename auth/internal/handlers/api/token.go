package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) exchangeToken(c fiber.Ctx) error {
	var req models.ExchangeTokenRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	ctx := c.Context()
	if err := req.ValidateAuthorization(ctx); err != nil {
		return responseutil.SendError(c, errors.ErrInvalidParams.WithError(err))
	}
	res, err := h.tokenService.ExchangeToken(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, res)
}

func (h *handler) getCurrentAccessToken(c fiber.Ctx) error {
	ctx := c.Context()
	res, err := h.tokenService.GetCurrentAccessToken(ctx)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, res)
}

func (h *handler) revokeCurrentToken(c fiber.Ctx) error {
	var req models.RevokeCurrentTokenRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	ctx := c.Context()
	req.ParseJWT(ctx, h.defaultConfig) // Parse JWT from Authorization header if present
	err = h.tokenService.RevokeCurrentToken(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
