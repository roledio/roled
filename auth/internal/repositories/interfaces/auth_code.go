package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type AuthCodeRepository interface {
	Create(ctx context.Context, authCode *entities.AuthCode) error
	FindByClientIDAndCodeHash(ctx context.Context, clientID string, codeHash string) (*entities.AuthCode, error)
	UpdateUsedAuthCode(ctx context.Context, authCode *entities.AuthCode) (int, error)
}
