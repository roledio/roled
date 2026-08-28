package project

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/services/upload"
	"github.com/roledio/roled/pkg/errors"
	"golang.org/x/sync/singleflight"
)

type ProjectService interface {
	GetConsoleConfig(ctx context.Context) (*models.GetConsoleConfigResponse, error)
	GetProjects(ctx context.Context, req *models.GetProjectsRequest) ([]models.GetProjectsResponse, int, error)
	GetProjectDetails(ctx context.Context, req *models.GetProjectDetailsRequest) (*models.ProjectDetails, error)
	CreateProject(ctx context.Context, req *models.CreateProjectRequest) (*models.ProjectDetails, error)
	UpdateProject(ctx context.Context, req *models.UpdateProjectRequest) (*models.ProjectDetails, error)
	DeleteProject(ctx context.Context, req *models.DeleteProjectRequest) error
	GetProjectSettings(ctx context.Context, req *models.GetProjectSettingsRequest) (*models.ProjectSettings, error)
	UpdateProjectSettings(ctx context.Context, req *models.UpdateProjectSettingsRequest) (*models.ProjectSettings, error)
	UpdateProjectSignupRole(ctx context.Context, req *models.UpdateProjectSignupRoleRequest) (*models.UpdateProjectSignupRoleResponse, error)
}

type projectService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	uploadService upload.UploadService
	uploadBaseURL string
	redis         infra.RedisService

	// sfGroup is a zero-value singleflight.Group. In Go, singleflight.Group zero-value is immediately
	// valid and thread-safe, so it does not need to be explicitly initialized in NewProjectService.
	sfGroup singleflight.Group
}

func NewProjectService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, uploadService upload.UploadService, redis infra.RedisService) ProjectService {
	var uploadBaseURL string
	switch defaultConfig.Upload.Driver {
	case constants.UploadDriverLocal:
		uploadBaseURL = defaultConfig.BaseURL + "/uploads"
	case constants.UploadDriverS3:
		uploadBaseURL = defaultConfig.Upload.S3.BaseURL
	}
	return &projectService{
		defaultConfig: defaultConfig,
		registry:      registry,
		uploadService: uploadService,
		uploadBaseURL: uploadBaseURL,
		redis:         redis,
	}
}

func (s *projectService) GetConsoleConfig(ctx context.Context) (*models.GetConsoleConfigResponse, error) {
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindSystem(ctx)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find system project", "error", err)
		return nil, errors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Error("System project not found")
		return nil, errors.ErrSystemError.WithError(fmt.Errorf("system project not found"))
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("System project is not active", "project_id", project.ID)
		return nil, errors.ErrSystemError.WithError(fmt.Errorf("system project is not active"))
	}
	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByProjectIDAndIsDefault(ctx, project.ID, true)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by project ID and is default", "error", err)
		return nil, errors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found by project ID and is default", "projectID", project.ID)
		return nil, errors.ErrSystemError.WithError(
			fmt.Errorf("default client not found for system project"))
	}
	if !client.IsActive {
		log.WithContext(ctx).Errorw("Client is not active", "client_id", client.ID)
		return nil, errors.ErrSystemError.WithError(
			fmt.Errorf("default client for system project is not active"))
	}
	res := models.GetConsoleConfigResponse{
		ClientID: client.ID,
	}
	return &res, nil
}
