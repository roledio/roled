package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gookit/goutil/strutil"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/errors"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories"
	"github.com/roledio/roled/internal/utils/contextutil"
	pkgconstants "github.com/roledio/roled/pkg/constants"
	pkgerrors "github.com/roledio/roled/pkg/errors"
	"github.com/roledio/roled/pkg/utils/encryptionutil"
	"github.com/roledio/roled/pkg/utils/idutil"
	"github.com/samber/lo"
	"go.openly.dev/pointy"
)

func (s *projectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest) (*models.ProjectDetails, error) {
	// Get current account from context, should not be nil here
	account := contextutil.GetAccount(ctx)
	if account == nil {
		return nil, errors.ErrCtxAccountNotFound
	}

	var tmpLogoURL *string
	if req.LogoURL != nil {
		tmpLogoURL = req.LogoURL
	}
	newLogoURL, isTmp := s.checkUploadLogoURL(req.LogoURL)

	// Find duplicate redirect URIs in the request
	mapRedirectURIs := make(map[string]*string)
	for _, redirectURI := range req.RedirectURIs {
		// It will always use the last login URL if there are duplicate redirect URIs
		mapRedirectURIs[redirectURI.RedirectURI] = lo.EmptyableToPtr(redirectURI.LoginURL)
	}

	var project *entities.Project
	err := s.registry.Tx(func(registry repositories.Registry) error {

		// Create project
		project = &entities.Project{
			ID:          idutil.NewID(),
			AccountID:   account.ID,
			Name:        req.Name,
			Description: req.Description,
			LogoURL:     newLogoURL,
			IsActive:    true,
			IsSystem:    false,
		}
		err := registry.ProjectRepository().Create(ctx, project)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create project", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create redirect URIs for the new project
		redirectURIs := make([]entities.RedirectURI, 0, len(mapRedirectURIs))
		for redirectURI, loginURL := range mapRedirectURIs {
			redirectURIs = append(redirectURIs, entities.RedirectURI{
				ProjectID:   project.ID,
				RedirectURI: redirectURI,
				LoginURL:    loginURL,
			})
		}
		err = registry.RedirectURIRepository().Create(ctx, redirectURIs)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create redirect URIs", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Create default project setting for the new project
		projectSetting := &entities.ProjectSetting{
			ID:                      idutil.NewID(),
			ProjectID:               project.ID,
			IsSignupEnabled:         false,
			DefaultSignupRoleID:     nil,
			IsSignupVerifyEmail:     false,
			IsForgotPasswordEnabled: false,
			IsAllowTempEmail:        false,
		}
		err = registry.ProjectSettingRepository().Create(ctx, projectSetting)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create project setting", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Generate client secret
		secretEncrypted, err := s.generateClientSecret(ctx)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to generate client secret", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}
		client := &entities.Client{
			ID:              idutil.NewID(),
			AccountID:       account.ID,
			ProjectID:       project.ID,
			Name:            "Main Client",
			Description:     pointy.String("The main client for " + project.Name),
			SecretEncrypted: secretEncrypted,
			IsActive:        true,
			IsDefault:       true,
		}

		// Create default client for the project
		err = registry.ClientRepository().Create(ctx, client)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create default client", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Prepare default resources and permissions for the new project
		resources, permissions, clientPermissions := s.buildResourcesAndPermissions(account.ID, project.ID, client.ID)

		// Insert resources
		_, err = registry.ResourceRepository().Create(ctx, resources)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create resources", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Insert permissions
		_, err = registry.PermissionRepository().Create(ctx, permissions)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create permissions", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Insert client permissions
		err = registry.ClientPermissionRepository().Create(ctx, clientPermissions)
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to create client permissions", "error", err)
			return pkgerrors.ErrSystemError.WithError(err)
		}

		// Make sure to move the logo file after the project is successfully created
		// If it is already moved and the transaction fails, the file cannot be used
		// again using the same logo URL
		if isTmp && tmpLogoURL != nil {
			err = s.moveFileFromTmp(ctx, *tmpLogoURL)
			if err != nil {
				// The error should rollback the transaction since the file must be successfully moved
				// or the logo URL will not be able to be accessed using the new URL
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result := models.ProjectDetails{
		ID:          project.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        project.Name,
		Description: project.Description,
		LogoURL:     project.LogoURL,
		IsActive:    project.IsActive,
	}

	for redirectURI, loginURL := range mapRedirectURIs {
		result.RedirectURIs = append(result.RedirectURIs, models.RedirectURI{
			RedirectURI: redirectURI,
			LoginURL:    lo.FromPtr(loginURL),
		})
	}

	return &result, nil
}

func (s *projectService) checkUploadLogoURL(logoURL *string) (*string, bool) {
	if logoURL == nil {
		return nil, false
	}
	tmpUploadURL := s.uploadBaseURL + "/tmp"
	if after, ok := strings.CutPrefix(*logoURL, tmpUploadURL); ok {
		newURL := s.uploadBaseURL + after // Return new logo URL without /tmp prefix
		return &newURL, true
	}
	return logoURL, false
}

func (s *projectService) moveFileFromTmp(ctx context.Context, logoURL string) error {
	// Extract the file path after /tmp
	tmpIndex := strings.Index(logoURL, "tmp/")
	if tmpIndex == -1 {
		log.WithContext(ctx).Errorw("Invalid logoURL: tmp/ not found", "logoURL", logoURL)
		return pkgerrors.ErrSystemError.WithError(fmt.Errorf("invalid logoURL: tmp/ not found"))
	}
	tmpFilePath := logoURL[tmpIndex:]
	newFilePath := strings.TrimPrefix(tmpFilePath, "tmp/")

	log.WithContext(ctx).Debugw("Moving file from tmp", "from", tmpFilePath, "to", newFilePath)

	// Move the file using upload service
	err := s.uploadService.Move(ctx, tmpFilePath, newFilePath)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to move file from tmp", "error", err, "from", tmpFilePath, "to", newFilePath)
		return errors.ErrMoveTmpProjectLogo.WithError(err)
	}
	log.WithContext(ctx).Debugw("File moved successfully from tmp", "from", tmpFilePath, "to", newFilePath)
	return nil
}

func (s *projectService) getDefaultResourceMapping() map[string][]string {
	return map[string][]string{
		constants.ResourceCodeAccounts: {constants.ActionRead},
		constants.ResourceCodeProjects: {constants.ActionRead},
		constants.ResourceCodeClients: {
			constants.ActionRead,
			constants.ActionCreate,
			constants.ActionUpdate,
			constants.ActionDelete,
		},
		constants.ResourceCodeResources: {
			constants.ActionRead,
			constants.ActionCreate,
			constants.ActionUpdate,
			constants.ActionDelete,
		},
		constants.ResourceCodePermissions: {
			constants.ActionRead,
			constants.ActionCreate,
			constants.ActionUpdate,
			constants.ActionDelete,
		},
		constants.ResourceCodeUsers: {
			constants.ActionRead,
			constants.ActionCreate,
			constants.ActionUpdate,
			constants.ActionDelete,
		},
		constants.ResourceCodeRoles: {
			constants.ActionRead,
			constants.ActionCreate,
			constants.ActionUpdate,
			constants.ActionDelete,
		},
	}
}

func (s *projectService) buildResourcesAndPermissions(accountID, projectID, clientID string) ([]entities.Resource,
	[]entities.Permission, []entities.ClientPermission) {
	mapping := s.getDefaultResourceMapping()
	resources := []entities.Resource{}
	permissions := []entities.Permission{}
	for resourceCode, actions := range mapping {
		resourceName := strutil.UpperFirst(resourceCode)
		resource := entities.Resource{
			ID:          idutil.NewID(),
			AccountID:   accountID,
			ProjectID:   projectID,
			Code:        resourceCode,
			Name:        resourceName,
			Description: pointy.String(fmt.Sprintf("%s resource", resourceName)),
			IsDefault:   true,
		}
		resources = append(resources, resource)

		for _, actionCode := range actions {
			actionName := strutil.UpperFirst(actionCode)
			permission := entities.Permission{
				ID:          idutil.NewID(),
				ResourceID:  resource.ID,
				Code:        actionCode,
				Name:        actionName,
				Description: pointy.String(fmt.Sprintf("%s %s", actionName, resourceName)),
				IsDefault:   true,
			}
			permissions = append(permissions, permission)
		}
	}
	clientPermissions := []entities.ClientPermission{}
	for _, permission := range permissions {
		clientPermission := entities.ClientPermission{
			ClientID:     clientID,
			PermissionID: permission.ID,
		}
		clientPermissions = append(clientPermissions, clientPermission)
	}
	return resources, permissions, clientPermissions
}

func (s *projectService) generateClientSecret(ctx context.Context) (string, error) {
	purpose := pkgconstants.KeyPurposeClientSecret
	derivedKey, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to derive key for client secret encryption", "error", err)
		return "", err
	}
	secret := idutil.NanoID(64)
	secretEncrypted, err := encryptionutil.EncryptAES(secret, derivedKey, purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to encrypt client secret", "error", err)
		return "", err
	}
	return secretEncrypted, nil
}
