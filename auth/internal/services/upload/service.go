package upload

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/errors"
)

type UploadService interface {
	Upload(ctx context.Context, req *models.UploadFileRequest) (*models.UploadFileResponse, error)
	Delete(ctx context.Context, path string) error
	Move(ctx context.Context, from string, to string) error
}

func NewUploadService(defaultConfig *configs.DefaultConfig) UploadService {
	switch defaultConfig.Upload.Driver {
	case constants.UploadDriverLocal:
		return newLocalUploadService(defaultConfig)
	case constants.UploadDriverS3:
		return newS3UploadService(defaultConfig)
	default:
		panic("unsupported upload driver: " + defaultConfig.Upload.Driver)
	}
}

// validateContent validates the uploaded file content based on the upload type and size limit
func validateContent(ctx context.Context, req *models.UploadFileRequest, maxFileSize int64) (string, string, error) {
	// Check file size
	if req.File.Size > maxFileSize {
		log.WithContext(ctx).Warnw("Uploaded file size exceeds the allowed limit", "file_size", req.File.Size, "max_file_size", maxFileSize)
		return "", "", errors.ErrFileSizeTooLarge
	}

	// Validate the content based on upload type
	validator, err := newUploadTypeValidator(req.Type)
	if err != nil {
		log.WithContext(ctx).Warnw("Invalid upload type", "error", err)
		return "", "", err
	}
	ext, contentType, err := validator.Validate(ctx, req)
	if err != nil {
		log.WithContext(ctx).Warnw("Upload type validation failed", "error", err)
		return "", "", err
	}

	// Generate UUID filename with extension
	filename := uuid.NewString() + ext

	return filename, contentType, nil
}
