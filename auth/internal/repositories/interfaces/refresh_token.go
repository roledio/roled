package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type RefreshTokenRepository interface {
	FindByClientIDAndRefreshTokenHash(ctx context.Context, clientID, refreshTokenHash string) (*entities.RefreshToken, error)
	UpdateUsedRefreshToken(ctx context.Context, refreshToken *entities.RefreshToken) (int, error)
	UpdateAsRevoked(ctx context.Context, refreshToken *entities.RefreshToken) (int, error)
	Create(ctx context.Context, refreshToken *entities.RefreshToken) error
}
