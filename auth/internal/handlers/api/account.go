package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) getCurrentAccountDetails(c fiber.Ctx) error {
	ctx := c.Context()
	accessToken := c.Locals(constants.CtxAccessToken).(*entities.AccessToken)
	if accessToken == nil {
		log.WithContext(ctx).Error("Access token not found in locals")
		return responseutil.SendError(c, errors.ErrSystemError)
	}
	req := models.GetAccountDetailsRequest{
		AccountID: accessToken.AccountID,
	}
	account, err := h.accountService.GetAccountDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, account)
}

func (h *handler) getAccountDetails(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetAccountDetailsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	account, err := h.accountService.GetAccountDetails(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, account)
}

func (h *handler) getAccounts(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.GetAccountsRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	accounts, total, err := h.accountService.GetAccounts(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	pagination := responseutil.Paginate(req.PageRequest, len(accounts), total)
	return responseutil.SendSuccessWithPagination(c, accounts, pagination)
}

func (h *handler) updateAccount(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.UpdateAccountRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	account, err := h.accountService.UpdateAccount(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, account)
}

func (h *handler) deleteAccount(c fiber.Ctx) error {
	ctx := c.Context()
	var req models.DeleteAccountRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	err = h.accountService.DeleteAccount(ctx, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, nil)
}
