package authorize

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *authorizeService) RenderAuthorize(ctx context.Context, req *models.RenderAuthorizeRequest) (*models.RenderAuthorizeResult, error) {
	project, _, projectSetting, err := s.validateAuthorizeRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	res := models.RenderAuthorizeResult{
		Project:        project,
		ProjectSetting: projectSetting,
	}
	if projectSetting.IsForgotPasswordEnabled {
		// Build forgot password URL when enabled
		res.ForgotPasswordURLPath = s.buildForgotPasswordPath(req)
	}
	conns, err := s.registry.OAuthConnectionRepository().FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find OAuth connections by project ID", "project_id", project.ID, "error", err)
		return &res, nil
	}
	// Build oauth connection URLs when enabled
	s.buildOAuthConnections(conns, req, &res)
	return &res, nil
}

func (s *authorizeService) validateAuthorizeRequest(ctx context.Context, req *models.RenderAuthorizeRequest) (
	*entities.Project, *entities.RedirectURI, *entities.ProjectSetting, error) {
	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by ID", "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found by ID", "client_id", req.ClientID)
		return nil, nil, nil, errors.ErrInvalidClientID
	}
	if !client.IsActive {
		log.WithContext(ctx).Errorw("Client is not active", "client_id", req.ClientID)
		return nil, nil, nil, errors.ErrClientNotActive
	}
	accountRepo := s.registry.AccountRepository()
	account, err := accountRepo.FindByID(ctx, client.AccountID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find account by client ID", "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if account == nil {
		log.WithContext(ctx).Errorw("Account not found by client ID", "client_id", req.ClientID)
		return nil, nil, nil, errors.ErrInvalidClientID
	}
	if !account.IsActive {
		log.WithContext(ctx).Errorw("Account is not active", "account_id", account.ID)
		return nil, nil, nil, errors.ErrInvalidClientID
	}
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, client.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by client ID", "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found by client ID", "client_id", req.ClientID)
		return nil, nil, nil, errors.ErrInvalidClientID
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "project_id", project.ID)
		return nil, nil, nil, errors.ErrProjectNotActive
	}
	redirectURI, err := s.registry.RedirectURIRepository().FindByProjectIDAndRedirectURI(ctx, project.ID, req.RedirectURI)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find redirect URI by project ID and redirect URI", "project_id", project.ID, "redirect_uri", req.RedirectURI, "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if redirectURI == nil {
		log.WithContext(ctx).Errorw("Redirect URI not found by project ID and redirect URI", "project_id", project.ID, "redirect_uri", req.RedirectURI)
		return nil, nil, nil, errors.ErrInvalidRedirectURI
	}
	projectSettingRepo := s.registry.ProjectSettingRepository()
	projectSetting, err := projectSettingRepo.FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project settings by project ID", "project_id", project.ID, "error", err)
		return nil, nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	return project, redirectURI, projectSetting, nil
}

func (s *authorizeService) buildOAuthConnections(
	conns []entities.OAuthConnection,
	req *models.RenderAuthorizeRequest,
	res *models.RenderAuthorizeResult) {
	for _, conn := range conns {
		if !conn.Enabled {
			continue
		}
		soc := models.OAuthConnection{
			Provider: conn.Provider,
		}
		if conn.Provider == constants.OAuthProviderGoogle {
			soc.URLPath = s.buildGoogleOAuthPath(req)
		}
		res.OAuthConnections = append(res.OAuthConnections, soc)
	}
}

func (s *authorizeService) buildGoogleOAuthPath(req *models.RenderAuthorizeRequest) string {
	query := url.Values{}
	query.Set("client_id", req.ClientID)
	query.Set("redirect_uri", req.RedirectURI)
	query.Set("response_type", req.ResponseType)
	query.Set("code_challenge", req.CodeChallenge)
	query.Set("code_challenge_method", req.CodeChallengeMethod)
	query.Set("state", req.State)
	return fmt.Sprintf("/oauth/google?%s", query.Encode())
}

func (s *authorizeService) buildForgotPasswordPath(req *models.RenderAuthorizeRequest) string {
	query := url.Values{}
	query.Set("client_id", req.ClientID)
	query.Set("redirect_uri", req.RedirectURI)
	return fmt.Sprintf("/password/forgot?%s", query.Encode())
}
