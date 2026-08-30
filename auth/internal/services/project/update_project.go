package project

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/shared"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/samber/lo"
)

func (s *projectService) UpdateProject(ctx context.Context, req *models.UpdateProjectRequest) (*models.ProjectDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.IsSystem {
		log.WithContext(ctx).Errorw("Attempt to modify system project", "project_id", req.ProjectID)
		return nil, errors.ErrModifySystemProject
	}

	var tmpLogoURL *string
	if req.LogoURL != nil {
		tmpLogoURL = req.LogoURL
	}
	newLogoURL, isTmpLogoURL := s.checkUploadLogoURL(req.LogoURL)
	oldLogoURL := project.LogoURL

	// Find duplicate redirect URIs in the request
	mapRedirectURIs := make(map[string]*string)
	for _, redirectURI := range req.RedirectURIs {
		// It will always use the last login URL if there are duplicate redirect URIs
		mapRedirectURIs[redirectURI.RedirectURI] = lo.EmptyableToPtr(redirectURI.LoginURL)
	}

	var oldRedirectURIs []entities.RedirectURI
	err = s.registry.Tx(func(registry repositories.Registry) error {
		project.Name = req.Name
		project.Description = req.Description
		project.LogoURL = newLogoURL
		project.IsActive = *req.IsActive
		affected, err := registry.ProjectRepository().Update(ctx, project)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to update project", "error", err, "project_id", req.ProjectID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		if affected == 0 {
			log.WithContext(ctx).Errorw("No project updated", "project_id", req.ProjectID)
			return errors.ErrProjectNotFound
		}

		redirectURIRepository := registry.RedirectURIRepository()
		// Get old redirect URIs for cache invalidation
		oldRedirectURIs, err = redirectURIRepository.FindByProjectID(ctx, project.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to get redirect URIs by project ID", "error", err, "project_id", project.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		// Update redirect URIs: delete all existing redirect URIs and create new ones
		_, err = redirectURIRepository.DeleteByProjectID(ctx, project.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to delete existing redirect URIs by project ID", "error", err, "project_id", project.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		var redirectURIs []entities.RedirectURI
		for redirectURI, loginURL := range mapRedirectURIs {
			redirectURIs = append(redirectURIs, entities.RedirectURI{
				ProjectID:   project.ID,
				RedirectURI: redirectURI,
				LoginURL:    loginURL,
			})
		}
		err = redirectURIRepository.Create(ctx, redirectURIs)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create redirect URIs", "error", err, "project_id", project.ID)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Make sure to move the logo file after the project is successfully updated
		// If it is already moved and the transaction fails, the file cannot be used
		// again using the same logo URL
		if isTmpLogoURL && tmpLogoURL != nil {
			err = s.moveFileFromTmp(ctx, *tmpLogoURL)
			if err != nil {
				// The error should rollback the transaction since the file must be successfully moved
				// or the logo URL will not be able to be accessed using the new URL
				return err
			}
		}

		// If the logo URL is changed (or become empty) and the old logo URL is not nil, delete the old logo file
		if oldLogoURL != nil && (newLogoURL == nil || *oldLogoURL != *newLogoURL) {
			err = s.deleteFile(ctx, *oldLogoURL)
			if err != nil {
				log.WithContext(ctx).Errorw("Failed to delete old logo file", "error", err, "logo_url", *oldLogoURL)
				// The error should not rollback the transaction since the project has been successfully updated, the old logo file can be deleted later
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cache after successful update
	shared.InvalidateProjectCache(ctx, s.redis, project)
	shared.InvalidateRedirectURICache(ctx, s.redis, project.ID, oldRedirectURIs)

	res := &models.ProjectDetails{
		ID:          project.ID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
		Name:        project.Name,
		Description: project.Description,
		LogoURL:     project.LogoURL,
		IsActive:    project.IsActive,
	}
	for redirectURI, loginURL := range mapRedirectURIs {
		res.RedirectURIs = append(res.RedirectURIs, models.RedirectURI{
			RedirectURI: redirectURI,
			LoginURL:    lo.FromPtr(loginURL),
		})
	}
	return res, nil
}

func (s *projectService) deleteFile(ctx context.Context, logoURL string) error {
	// The logo URL is expected to be in the format of {uploadBaseURL}/{filePath}, so we need to trim the uploadBaseURL to get the file path
	filePath := strings.TrimPrefix(logoURL, s.uploadBaseURL+"/")
	if filePath == logoURL { // The file path is the same as the logo URL, unable to trim the upload base URL prefix
		log.WithContext(ctx).Debugw("Logo URL does not contain upload base URL prefix", "logo_url", logoURL, "upload_base_url", s.uploadBaseURL)
		return nil // The logo URL does not contain the expected prefix, so we cannot determine the file path to delete. Log the error and return nil to avoid blocking the main flow.
	}
	err := s.uploadService.Delete(ctx, filePath)
	if err != nil {
		return err
	}
	log.WithContext(ctx).Debugw("File deleted successfully", "file_path", filePath)
	return nil
}
