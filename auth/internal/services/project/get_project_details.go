package project

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/samber/lo"
)

func (s *projectService) GetProjectDetails(ctx context.Context, req *models.GetProjectDetailsRequest) (*models.ProjectDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}

	redirectURIs, err := s.registry.RedirectURIRepository().FindByProjectID(ctx, req.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find redirect URIs by project ID", "error", err, "project_id", project.ID)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}

	resp := &models.ProjectDetails{
		ID:          project.ID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		Name:        project.Name,
		Description: project.Description,
		LogoURL:     project.LogoURL,
		IsActive:    project.IsActive,
	}
	for _, redirectURI := range redirectURIs {
		resp.RedirectURIs = append(resp.RedirectURIs, models.RedirectURI{
			RedirectURI: redirectURI.RedirectURI,
			LoginURL:    lo.FromPtr(redirectURI.LoginURL),
		})
	}
	return resp, nil
}
