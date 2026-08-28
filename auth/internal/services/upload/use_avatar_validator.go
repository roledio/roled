package upload

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/errors"
)

const maxAvatarSize = 2 * 1024 * 1024 // 2 MB

type userAvatarValidator struct {
}

func (v *userAvatarValidator) Validate(ctx context.Context, req *models.UploadFileRequest) (string, string, error) {
	// Check file size first
	if req.File.Size > maxAvatarSize {
		log.WithContext(ctx).Errorw("Avatar size exceeds max size", "size", req.File.Size, "max_size", maxAvatarSize)
		return "", "", errors.ErrFileSizeTooLarge
	}

	// Validate image content
	ext, contentType, err := validateImage(ctx, req)
	if err != nil {
		return "", "", err
	}

	return ext, contentType, nil
}
