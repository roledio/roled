package user

import (
	"context"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/services/upload"
)

type UserService interface {
	RenderVerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) (*models.VerifyEmailResult, error)
	RenderForgotPassword(ctx context.Context, req *models.RenderForgotPasswordRequest) (*models.RenderForgotPasswordResult, error)
	SubmitForgotPassword(ctx context.Context, req *models.SubmitForgotPasswordRequest) error
	RenderResetPassword(ctx context.Context, req *models.RenderResetPasswordRequest) (*models.RenderResetPasswordResult, error)
	SubmitResetPassword(ctx context.Context, req *models.SubmitResetPasswordRequest) (*models.SubmitResetPasswordResult, error)
	GetUsers(ctx context.Context, req *models.GetUsersRequest) ([]models.UserDetails, int, error)
	GetUserDetails(ctx context.Context, req *models.GetUserDetailsRequest) (*models.UserDetails, error)
	GetCurrentUserDetails(ctx context.Context, includePermissions bool) (*models.UserDetails, error)
	GetExternalUserDetails(ctx context.Context, req *models.GetExternalUserDetailsRequest) (*models.UserDetails, error)
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserDetails, error)
	UpdateUser(ctx context.Context, req *models.UpdateUserRequest) (*models.UserDetails, error)
	UpdateCurrentUser(ctx context.Context, req *models.UpdateCurrentUserRequest) (*models.UserDetails, error)
	DeleteUser(ctx context.Context, req *models.DeleteUserRequest) error
	ResendVerificationEmail(ctx context.Context, req *models.ResendVerificationEmailRequest) error
	RequestPasswordReset(ctx context.Context, req *models.RequestPasswordResetRequest) error
}

type userService struct {
	defaultConfig  *configs.DefaultConfig
	registry       repositories.Registry
	redisService   infra.RedisService
	emailPublisher queues.Publisher
	uploadService  upload.UploadService
	uploadBaseURL  string
}

func NewUserService(
	defaultConfig *configs.DefaultConfig,
	registry repositories.Registry,
	uploadService upload.UploadService,
	redisService infra.RedisService,
	emailPublisher queues.Publisher) UserService {

	var uploadBaseURL string
	switch defaultConfig.Upload.Driver {
	case constants.UploadDriverLocal:
		uploadBaseURL = defaultConfig.BaseURL + "/uploads"
	case constants.UploadDriverS3:
		uploadBaseURL = defaultConfig.Upload.S3.BaseURL
	}
	return &userService{
		defaultConfig:  defaultConfig,
		registry:       registry,
		redisService:   redisService,
		emailPublisher: emailPublisher,
		uploadService:  uploadService,
		uploadBaseURL:  uploadBaseURL,
	}
}
