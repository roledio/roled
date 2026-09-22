package branding

import (
	"context"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/upload"
)

type Service interface {
	GetBranding(context.Context, *models.GetBrandingRequest) (*models.BrandingDetails, error)
	UpdateBranding(context.Context, *models.UpdateBrandingRequest) (*models.BrandingDetails, error)
	ResolveBranding(context.Context, string) (*models.BrandingDetails, error)
}

type service struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	uploadService upload.UploadService
	uploadBaseURL string
}

func NewService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, uploadService upload.UploadService) Service {
	var uploadBaseURL string
	switch defaultConfig.Upload.Driver {
	case constants.UploadDriverLocal:
		uploadBaseURL = defaultConfig.BaseURL + "/uploads"
	case constants.UploadDriverS3:
		uploadBaseURL = defaultConfig.Upload.S3.BaseURL
	}
	return &service{
		defaultConfig: defaultConfig,
		registry:      registry,
		uploadService: uploadService,
		uploadBaseURL: uploadBaseURL,
	}
}
