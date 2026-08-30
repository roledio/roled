package user

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *userService) RenderForgotPassword(ctx context.Context, req *models.RenderForgotPasswordRequest) (*models.RenderForgotPasswordResult, error) {
	project, loginURL, err := s.validateForgotPassword(ctx, req.ClientID, req.RedirectURI)
	if err != nil {
		return nil, err
	}
	result := &models.RenderForgotPasswordResult{
		Project:  project,
		LoginURL: loginURL,
	}
	return result, nil
}

func (s *userService) validateForgotPassword(ctx context.Context, clientID, redirectURI string) (*entities.Project, *string, error) {
	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByID(ctx, clientID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by client ID", "error", err, "client_id", clientID)
		return nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found", "client_id", clientID)
		return nil, nil, errors.ErrClientNotFound
	}
	if !client.IsActive {
		log.WithContext(ctx).Errorw("Client is not active", "client_id", clientID)
		return nil, nil, errors.ErrClientNotActive
	}
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, client.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by client ID", "error", err, "client_id", clientID)
		return nil, nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found", "client_id", clientID)
		return nil, nil, errors.ErrProjectNotFound
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "client_id", clientID)
		return nil, nil, errors.ErrProjectNotActive
	}

	uri := strings.TrimSpace(redirectURI)
	var loginURL *string
	if uri != "" {
		redirectURI, err := s.registry.RedirectURIRepository().FindByProjectIDAndRedirectURI(ctx, project.ID, uri)
		if err != nil {
			log.WithContext(ctx).Warnw("Failed to find redirect URI by project ID and redirect URI", "error", err, "project_id", project.ID, "redirect_uri", uri)
			// We will not return error here, as we want to allow users to proceed even if the redirect URI is not found.
			// The login URL will simply be nil in that case.
		}
		if redirectURI != nil {
			loginURL = redirectURI.LoginURL
		}
	}

	return project, loginURL, nil
}
