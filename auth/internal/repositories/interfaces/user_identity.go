package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type UserIdentityRepository interface {
	Create(ctx context.Context, userIdentity *entities.UserIdentity) error
	FindByProviderAndProviderUserIDAndProjectID(ctx context.Context, provider, providerUserID, projectID string) (*entities.UserIdentity, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.UserIdentity, error)
	DeleteByID(ctx context.Context, id string) (int, error)
}
