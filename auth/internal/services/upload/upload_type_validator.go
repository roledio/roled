package upload

import (
	"context"
	"image"
	"io"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
)

var (
	allowedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
	}
)

type UploadTypeValidator interface {
	// Validate validates the uploaded content and returns the file extension and content type if valid
	Validate(ctx context.Context, req *models.UploadFileRequest) (string, string, error)
}

func newUploadTypeValidator(uploadType string) (UploadTypeValidator, error) {
	switch uploadType {
	case constants.UploadTypeProjectLogo:
		return &projectLogoValidator{}, nil
	case constants.UploadTypeUserAvatar:
		return &userAvatarValidator{}, nil
	default:
		return nil, errors.ErrInvalidUploadType
	}
}

func validateImage(ctx context.Context, req *models.UploadFileRequest) (string, string, error) {
	file, err := req.File.Open()
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to open uploaded file for validation", "error", err)
		return "", "", err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Errorw("Failed to close uploaded file", "error", err)
		}
	}()

	// Read first 512 bytes for content type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		log.WithContext(ctx).Errorw("Failed to read from file", "error", err)
		return "", "", err
	}
	buffer = buffer[:n]

	// Check content type
	contentType := http.DetectContentType(buffer)
	ext := allowedImageTypes[contentType]
	if ext == "" {
		log.WithContext(ctx).Errorw("Invalid image content type", "content_type", contentType)
		return "", "", errors.ErrInvalidImageType
	}

	// Reset file to beginning for decoding
	_, err = file.Seek(0, 0)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to seek file to beginning", "error", err)
		return "", "", err
	}

	// Decode to verify it is a valid image
	_, _, err = image.DecodeConfig(file)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to decode image", "error", err)
		return "", "", err
	}
	return ext, contentType, nil
}
