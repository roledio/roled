package member

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/services/infra"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

// MemberService defines the member service interface.
//
// A member is a user that belongs to the system project (Roled Console).
// Users in other projects are not considered members; they are only associated
// with their respective projects.
//
// This service is strictly scoped to the system project context, which means all
// operations performed in this service must always check and validate against the
// system project from the access token, except methods for the web handler.
type MemberService interface {
	GetMembers(ctx context.Context, req *models.GetMembersRequest) ([]models.GetMembersResponse, int, error)
	GetMemberDetails(ctx context.Context, req *models.GetMemberDetailsRequest) (*models.GetMemberDetailsResponse, error)
	CreateMember(ctx context.Context, req *models.CreateMemberRequest) (*models.CreateMemberResponse, error)
	UpdateMember(ctx context.Context, req *models.UpdateMemberRequest) (*models.UpdateMemberResponse, error)
	DeleteMember(ctx context.Context, req *models.DeleteMemberRequest) error

	// Used in web handler

	RenderActivateMember(ctx context.Context, req *models.RenderActivateMemberRequest) (*models.RenderActivateMemberResponse, error)
	SubmitActivateMember(ctx context.Context, req *models.SubmitActivateMemberRequest) (*models.SubmitActivateMemberResponse, error)
}

type memberService struct {
	defaultConfig  *configs.DefaultConfig
	registry       repositories.Registry
	emailPublisher queues.Publisher
	redisService   infra.RedisService
}

func NewMemberService(defaultConfig *configs.DefaultConfig, registry repositories.Registry,
	emailPublisher queues.Publisher, redisService infra.RedisService) MemberService {
	return &memberService{
		defaultConfig:  defaultConfig,
		registry:       registry,
		emailPublisher: emailPublisher,
		redisService:   redisService,
	}
}

// validateMustSystemProject validates that the access token in the context
// belongs to the system project. It returns the access token and system project
// entity if validation passes.
func (s *memberService) validateMustSystemProject(ctx context.Context) (*entities.AccessToken, *entities.Project, error) {
	projectRepo := s.registry.ProjectRepository()
	systemProject, err := projectRepo.FindSystem(ctx)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find system project", "error", err)
		return nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if systemProject == nil {
		log.WithContext(ctx).Errorw("System project not found")
		return nil, nil, pkgerrors.ErrSystemError.WithError(fmt.Errorf("system project not found"))
	}
	if !systemProject.IsActive {
		log.WithContext(ctx).Errorw("System project is not active", "project_id", systemProject.ID)
		return nil, nil, pkgerrors.ErrSystemError.WithError(fmt.Errorf("system project is not active"))
	}

	// Get current access token from context, should not be nil here
	accessToken := contextutil.GetAccessToken(ctx)
	if accessToken == nil {
		return nil, nil, errors.ErrCtxAccessTokenNotFound
	}

	// Validate that the access token belongs to the system project
	if accessToken.ProjectID != systemProject.ID {
		log.WithContext(ctx).Errorw("Project ID from JWT does not match system project ID", "jwt_project_id", accessToken.ProjectID, "system_project_id", systemProject.ID)
		return nil, nil, pkgerrors.ErrOperationNotAvailable
	}
	return accessToken, systemProject, nil
}
