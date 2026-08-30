package upload

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
)

const maxProjectLogoSize = 2 * 1024 * 1024 // 2 MB

type projectLogoValidator struct {
}

func (v *projectLogoValidator) Validate(ctx context.Context, req *models.UploadFileRequest) (string, string, error) {
	// Check file size first
	if req.File.Size > maxProjectLogoSize {
		log.WithContext(ctx).Errorw("Project logo size exceeds max size", "size", req.File.Size, "max_size", maxProjectLogoSize)
		return "", "", errors.ErrFileSizeTooLarge
	}

	// Validate image content
	ext, contentType, err := validateImage(ctx, req)
	if err != nil {
		return "", "", err
	}

	return ext, contentType, nil
}
