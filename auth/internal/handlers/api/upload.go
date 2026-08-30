package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/utils/requestutil"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

func (h *handler) uploadFile(c fiber.Ctx) error {
	var req models.UploadFileRequest
	err := requestutil.BindAndValidate(c, &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	res, err := h.uploadService.Upload(c.Context(), &req)
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, res)
}
