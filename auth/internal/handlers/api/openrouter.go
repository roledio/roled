package api

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/responseutil"
)

type extractImageJSONRequest struct {
	Image string `json:"image"`
	URL   string `json:"url"`
}

func (h *handler) extractImage(c fiber.Ctx) error {
	ctx := c.Context()

	// 1. Try to get uploaded image from multipart/form-data
	var file *multipart.FileHeader
	var err error

	file, err = c.FormFile("image")
	if err != nil {
		file, err = c.FormFile("file")
	}

	// If not found by name, check any uploaded file in multipart form
	if err != nil || file == nil {
		if form, formErr := c.MultipartForm(); formErr == nil && form != nil {
			for _, files := range form.File {
				if len(files) > 0 {
					file = files[0]
					err = nil
					break
				}
			}
		}
	}

	if file != nil {
		result, extractErr := h.openRouterService.ExtractInvoiceFromImage(ctx, file)
		if extractErr != nil {
			return responseutil.SendError(c, errors.ErrSystemError.WithError(extractErr))
		}
		if c.Query("raw") == "true" {
			return c.JSON(result)
		}
		return responseutil.SendSuccess(c, result)
	}

	// 2. Alternatively, check if request is JSON containing base64 data or image URL
	if strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		var req extractImageJSONRequest
		if bindErr := c.Bind().Body(&req); bindErr == nil {
			target := req.Image
			if target == "" {
				target = req.URL
			}
			if target != "" {
				result, extractErr := h.openRouterService.ExtractInvoiceFromBase64OrURL(ctx, target)
				if extractErr != nil {
					return responseutil.SendError(c, errors.ErrSystemError.WithError(extractErr))
				}
				if c.Query("raw") == "true" {
					return c.JSON(result)
				}
				return responseutil.SendSuccess(c, result)
			}
		}
	}

	return responseutil.SendError(c, errors.ErrInvalidParams.WithDebugMessage("No image file provided. Please upload an image using form-data with field name 'image' or 'file', or send a JSON body with 'image' base64 data."))
}
