package requestutil

import (
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/validationutil"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func BindAndValidate(c fiber.Ctx, req any) error {
	ctx := c.Context()
	_ = c.Bind().Header(req)
	_ = c.Bind().URI(req)
	_ = c.Bind().Query(req)
	_ = c.Bind().Body(req)
	err := validationutil.ValidateStruct(ctx, req)
	if err != nil {
		log.WithContext(ctx).Error("Failed to validate request struct: ", err)
		return errors.ErrInvalidParams.WithError(err)
	}
	return nil
}
