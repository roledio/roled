package responseutil

import (
	"errors"

	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/models"

	"github.com/gofiber/fiber/v3"
)

var DebugError bool

func SendSuccess(c fiber.Ctx, data interface{}) error {
	return SendSuccessWithPagination(c, data, nil)
}

func SendSuccessWithPagination(c fiber.Ctx, data interface{}, pagination *models.Pagination) error {
	body := models.ResponseBody{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}
	return c.JSON(body)
}

func SendError(c fiber.Ctx, err error) error {
	var ce pkgerrors.CustomError
	if !errors.As(err, &ce) {
		// This error is not of type pkgerrors.CustomError, it is an unexpected error
		ce = pkgerrors.ErrSystemError.WithError(err)
	}
	body := models.ResponseBody{
		Success: false,
		Error: &models.ErrorBody{
			Code:    ce.Code,
			Message: ce.Msg,
		},
	}
	if DebugError && ce.Err != nil {
		debug := ce.Err.Error()
		unwrap := errors.Unwrap(ce.Err)
		if unwrap != nil {
			debug = unwrap.Error() // Use the root error message for debug
		}
		body.Error.Debug = &debug
	} else if DebugError && ce.DebugMessage != "" {
		body.Error.Debug = &ce.DebugMessage
	}
	err = c.JSON(body)
	if err != nil {
		return err
	}
	return c.SendStatus(ce.HttpCode)
}

func Paginate(req models.PageRequest, actualSize, totalData int) *models.Pagination {
	if actualSize > req.PageSize || actualSize > totalData {
		panic("the actual size must not be greater than page size or total data")
	}
	pagination := models.Pagination{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		TotalData: totalData,
	}
	if req.PageSize > totalData {
		pagination.PageSize = totalData
	}
	if req.PageSize > actualSize {
		pagination.PageSize = actualSize
	}
	return &pagination
}

func IsSuccessful(c fiber.Ctx) bool {
	code := c.Response().StatusCode()
	return code >= 200 && code < 300
}
