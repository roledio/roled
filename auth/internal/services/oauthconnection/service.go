package oauthconnection

import (
	"context"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/services/infra"
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
	redisService  infra.RedisService
}

func NewOAuthConnectionService(defaultConfig *configs.DefaultConfig, registry repositories.Registry, redisService infra.RedisService) OAuthConnectionService {
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
