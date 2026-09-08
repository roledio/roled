package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type OAuthConnectionRepository interface {
	FindByProjectID(ctx context.Context, projectID string) ([]entities.OAuthConnection, error)
	FindByProjectIDAndProvider(ctx context.Context, projectID, provider string) (*entities.OAuthConnection, error)
	Create(ctx context.Context, connection *entities.OAuthConnection) error
	Update(ctx context.Context, connection *entities.OAuthConnection) (int, error)
	Delete(ctx context.Context, projectID, provider string) (int, error)
}
