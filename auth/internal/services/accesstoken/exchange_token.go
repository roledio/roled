package accesstoken

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/karrick/tparse/v2"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/utils/idutil"
	"github.com/roledio/roled/auth/pkg/utils/jwtutil"
	"github.com/shomali11/util/xhashes"
)

func (s *accessTokenService) ExchangeToken(ctx context.Context, req *models.ExchangeTokenRequest) (*models.ExchangeTokenResponse, error) {
	clientRepo := s.registry.ClientRepository()
	client, err := clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find client by ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if client == nil {
		log.WithContext(ctx).Errorw("Client not found by ID", "client_id", req.ClientID)
		return nil, errors.ErrInvalidClientID
	}
	if !client.IsActive {
		log.WithContext(ctx).Errorw("Client is not active", "client_id", req.ClientID)
		return nil, errors.ErrClientNotActive
	}
	// Validate project
	projectRepo := s.registry.ProjectRepository()
	project, err := projectRepo.FindByID(ctx, client.ProjectID)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to find project by ID", "error", err)
		return nil, pkgerrors.ErrSystemError.WithError(err)
	}
	if project == nil {
		log.WithContext(ctx).Errorw("Project not found by ID", "id", client.ProjectID)
		return nil, errors.ErrInvalidClientID
	}
	if !project.IsActive {
		log.WithContext(ctx).Errorw("Project is not active", "project_id", project.ID)
		return nil, errors.ErrProjectNotActive
	}
	switch req.GrantType {
	case constants.GrantTypeClientCredentials:
		return s.handleClientCredentials(ctx, req, client, project)
	case constants.GrantTypeRefreshToken:
		return s.handleRefreshToken(ctx, req, client, project)
	case constants.GrantTypeAuthorizationCode:
		return s.handleAuthorizationCode(ctx, req, client, project)
	}
	// The supported grant types should have been handled above
	return nil, errors.ErrInvalidGrantType
}

func (s *accessTokenService) buildAccessToken(ctx context.Context, grantType string, client *entities.Client, project *entities.Project,
	accountID string, userID *string) *entities.AccessToken {
	tokenID := idutil.NewID()
	issuedAt := time.Now()
	expiryDuration := s.parseAccessTokenExpiry(ctx, issuedAt)
	expiresIn := int(expiryDuration.Seconds())
	accessToken := entities.AccessToken{
		ID:        tokenID,
		AccountID: accountID,
		ProjectID: project.ID,
		ClientID:  client.ID,
		UserID:    userID,
		GrantType: grantType,
		IssuedAt:  &issuedAt,
		ExpiresIn: &expiresIn,
		Status:    constants.AccessTokenStatusIssued,
	}
	return &accessToken
}

func (s *accessTokenService) buildRefreshToken(ctx context.Context, client *entities.Client, project *entities.Project,
	accountID string, userID *string) (*entities.RefreshToken, string) {
	refreshTokenPlain := idutil.NanoID(64)
	refreshTokenHash := xhashes.SHA256(refreshTokenPlain)
	refreshTokenIssuedAt := time.Now()
	refreshTokenExpiryDuration := s.parseRefreshTokenExpiry(ctx, refreshTokenIssuedAt)
	refreshTokenExpiresIn := int(refreshTokenExpiryDuration.Seconds())
	return &entities.RefreshToken{
		ID:               idutil.NewID(),
		AccountID:        accountID,
		ProjectID:        project.ID,
		ClientID:         client.ID,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		Status:           constants.RefreshTokenStatusIssued,
		ExpiresIn:        &refreshTokenExpiresIn,
		IssuedAt:         &refreshTokenIssuedAt,
	}, refreshTokenPlain
}

func (s *accessTokenService) buildJWT(ctx context.Context, accessToken *entities.AccessToken) (string, error) {
	expiredAt := accessToken.IssuedAt.Add(time.Duration(*accessToken.ExpiresIn) * time.Second)
	claims := models.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessToken.ID,
			Subject:   accessToken.ClientID,
			Audience:  jwt.ClaimStrings{accessToken.ProjectID},
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			IssuedAt:  jwt.NewNumericDate(*accessToken.IssuedAt),
			Issuer:    s.defaultConfig.BaseURL,
		},
		ClientID: accessToken.ClientID,
	}
	if accessToken.UserID != nil {
		claims.UserID = *accessToken.UserID
		claims.Subject = *accessToken.UserID
	}
	jwt, err := jwtutil.GenerateToken(claims, s.defaultConfig.JWT.SigningKey)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to generate jwt access token", "error", err)
		return "", pkgerrors.ErrSystemError.WithError(err)
	}
	return jwt, nil
}

func (s *accessTokenService) parseAccessTokenExpiry(ctx context.Context, issuedAt time.Time) time.Duration {
	format := s.defaultConfig.JWT.ExpiryDuration
	duration, err := tparse.AbsoluteDuration(issuedAt, format)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to parse expiry duration, set default to 1 hour", "expiry_duration", format)
		duration = 1 * time.Hour // 1 hour
	}
	return duration
}

func (s *accessTokenService) parseRefreshTokenExpiry(ctx context.Context, issuedAt time.Time) time.Duration {
	format := s.defaultConfig.JWT.RefreshTokenExpiryDuration
	duration, err := tparse.AbsoluteDuration(issuedAt, format)
	if err != nil {
		log.WithContext(ctx).Warnw("Failed to parse refresh token expiry duration, set default to 7 days", "refresh_token_expiry_duration", format)
		duration = 7 * 24 * time.Hour // 7 days
	}
	return duration
}
