package project

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
)

func (s *projectService) DeleteProject(ctx context.Context, req *models.DeleteProjectRequest) error {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return err
	}
	if project.IsSystem {
		log.WithContext(ctx).Errorw("Attempt to delete system project", "project_id", req.ProjectID)
		return errors.ErrModifySystemProject
	}

	// Validate project name (case insensitive) in the request body to prevent accidental deletion
	if !strings.EqualFold(project.Name, req.Name) {
		log.WithContext(ctx).Errorw("Project name is missing or does not match existing project name", "project_id", req.ProjectID)
		return errors.ErrProjectNameRequiredForDeletion
	}

	// Get redirect URIs for this project to invalidate cache
	oldRedirectURIs, err := s.registry.RedirectURIRepository().FindByProjectID(ctx, project.ID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to get existing redirect URIs of project", "error", err, "project_id", req.ProjectID)
		return pkgerrors.ErrSystemError.WithError(err)
	}

	err = s.registry.Tx(func(registry repositories.Registry) error {

		// Delete all access tokens associated with this project
		_, err = registry.AccessTokenRepository().DeleteByProjectID(ctx, req.ProjectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete access tokens by project ID", "error", err, "project_id", req.ProjectID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Delete all clients associated with this project
		_, err = registry.ClientRepository().DeleteByProjectID(ctx, req.ProjectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete clients by project ID", "error", err, "project_id", req.ProjectID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Delete all users associated with this project
		_, err = registry.UserRepository().DeleteByProjectID(ctx, req.ProjectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete users by project ID", "error", err, "project_id", req.ProjectID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Finally, delete the project itself
		affected, err := registry.ProjectRepository().Delete(ctx, project)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete project by ID", "error", err, "project_id", req.ProjectID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No project deleted", "project_id", req.ProjectID)
			return errors.ErrProjectNotFound
		}

		oldLogoURL := project.LogoURL
		// If the logo URL is changed and the old logo URL is not nil, delete the old logo file
		if oldLogoURL != nil {
			err = s.deleteFile(ctx, *oldLogoURL)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to delete old logo file", "error", err, "logo_url", *oldLogoURL)
				// The error should not rollback the transaction since the project has been successfully updated and the old logo file can be deleted later
			}
		}
		return nil
	})

	if err == nil {
		// Invalidate caches related to this project after the transaction is committed successfully
		shared.InvalidateProjectCache(ctx, s.redis, project)
		shared.InvalidateProjectSettingCache(ctx, s.redis, project.ID)
		shared.InvalidateRedirectURICache(ctx, s.redis, project.ID, oldRedirectURIs)
	}

	return err
}
