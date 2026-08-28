package project

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

func (s *projectService) GetProjects(ctx context.Context, req *models.GetProjectsRequest) ([]models.GetProjectsResponse, int, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, 0, errors.ErrCtxAccountNotFound
	}
	projectRepo := s.registry.ProjectRepository()
	count, err := projectRepo.Count(ctx, req, account.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to count projects", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	if count == 0 {
		return nil, 0, nil
	}
	projects, err := projectRepo.FindAll(ctx, req, account.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find projects", "error", err)
		return nil, 0, pkgerrors.ErrSystemError.WithError(err)
	}
	var resp []models.GetProjectsResponse
	for _, project := range projects {
		resp = append(resp, models.GetProjectsResponse{
			ID:          project.ID,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			Name:        project.Name,
			Description: project.Description,
			LogoURL:     project.LogoURL,
			IsActive:    project.IsActive,
		})
	}
	return resp, count, nil
}
