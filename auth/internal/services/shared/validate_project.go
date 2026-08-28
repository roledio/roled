package shared

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgerrors "github.com/roledio/roled/pkg/errors"
)

// ValidateProject centralizes project validation logic for all services
func ValidateProject(ctx context.Context, registry repositories.Registry, projectID string) (*entities.Account, *entities.Project, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, nil, errors.ErrCtxAccountNotFound
	}

	var project *entities.Project
	var err error
	var projectRepo = registry.ProjectRepository()
	if account.IsSystem {
		project, err = projectRepo.FindByID(ctx, projectID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err, "project_id", projectID)
			return nil, nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if project == nil {
			log.WithContext(ctx).Errorw("Project not found by ID", "project_id", projectID)
			return nil, nil, errors.ErrProjectNotFound
		}
	} else {
		project, err = projectRepo.FindByIDAndAccountID(ctx, projectID, account.ID)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to find project by ID and account ID", "error", err, "project_id", projectID, "account_id", account.ID)
			return nil, nil, pkgerrors.ErrSystemError.WithError(err)
		}
		if project == nil {
			log.WithContext(ctx).Errorw("Project not found by ID and account ID",
				"project_id", projectID,
				"account_id", account.ID)
			return nil, nil, errors.ErrProjectNotFound
		}
	}
	return account, project, nil
}
