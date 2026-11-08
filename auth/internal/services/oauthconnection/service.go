package oauthconnection

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	pkgredis "github.com/roledio/roled/auth/pkg/redis"
	"github.com/roledio/roled/auth/pkg/utils/encryptionutil"
)

type OAuthConnectionService interface {
	GetOAuthConnections(ctx context.Context, req *models.GetOAuthConnectionsRequest) ([]models.OAuthConnectionDetails, error)
	GetOAuthConnectionDetails(ctx context.Context, req *models.GetOAuthConnectionRequest) (*models.OAuthConnectionDetails, error)
	CreateOAuthConnection(ctx context.Context, req *models.CreateOAuthConnectionRequest) (*models.OAuthConnectionDetails, error)
	UpdateOAuthConnection(ctx context.Context, req *models.UpdateOAuthConnectionRequest) (*models.OAuthConnectionDetails, error)
	DeleteOAuthConnection(ctx context.Context, req *models.DeleteOAuthConnectionRequest) error
}

type oAuthConnectionService struct {
	defaultConfig *configs.DefaultConfig
	registry      repositories.Registry
	redisService  pkgredis.Service
}

func NewOAuthConnectionService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redisService pkgredis.Service) OAuthConnectionService {
	return &oAuthConnectionService{
		defaultConfig: defaultConfig,
		registry:      registry,
		redisService:  redisService,
	}
}

func (s *oAuthConnectionService) encryptOAuthClientSecret(ctx context.Context, secret string) (string, error) {
	purpose := constants.KeyPurposeOAuthClientSecret
	derivedKey, err := encryptionutil.DeriveKey([]byte(s.defaultConfig.EncryptionMasterKey), purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to derive key for OAuth client secret encryption", "error", err)
		return "", err
	}
	secretEncrypted, err := encryptionutil.EncryptAES(secret, derivedKey, purpose)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to encrypt OAuth client secret", "error", err)
		return "", err
	}
	return secretEncrypted, nil
}

// A system OAuth connection must exist, but its enabled state does not affect default credentials.
func (s *oAuthConnectionService) validateSystemOAuthConnectionConfigured(ctx context.Context, projectID, provider string) error {
	systemProject, err := s.registry.ProjectRepository().FindSystem(ctx)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find system project", "error", err)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if systemProject == nil {
		log.WithContext(ctx).Warnw("System project does not exist", "project_id", projectID, "provider", provider)
		return pkgerrors.ErrSystemError.WithDebugMessage("System project does not exist")
	}

	systemConnection, err := s.registry.OAuthConnectionRepository().FindByProjectIDAndProvider(ctx, systemProject.ID, provider)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to check system OAuth connection", "error", err, "project_id", systemProject.ID, "provider", provider)
		return pkgerrors.ErrSystemError.WithError(err)
	}
	if systemConnection == nil {
		log.WithContext(ctx).Warnw("System project does not have OAuth connection configured for this provider", "project_id", projectID, "provider", provider)
		return errors.ErrDefaultCredentialsNotConfigured.WithDebugMessage("System OAuth connection is not configured for provider " + provider)
	}
	return nil
}

func (s *oAuthConnectionService) setCustomOAuthCredentials(ctx context.Context, connection *entities.OAuthConnection, clientID, clientSecret string, scopes []string) error {
	scopesStr := strings.Join(scopes, " ")
	secretEncrypted, err := s.encryptOAuthClientSecret(ctx, clientSecret)
	if err != nil {
		return err
	}

	connection.ClientID = &clientID
	connection.ClientSecretEncrypted = &secretEncrypted
	connection.Scopes = &scopesStr
	return nil
}
