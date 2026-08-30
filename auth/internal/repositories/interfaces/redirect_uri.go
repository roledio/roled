package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type RedirectURIRepository interface {
	FindByProjectIDAndRedirectURI(ctx context.Context, projectID string, redirectURI string) (*entities.RedirectURI, error)
	FindByProjectID(ctx context.Context, projectID string) ([]entities.RedirectURI, error)
	Create(ctx context.Context, redirectURIs []entities.RedirectURI) error
	DeleteByProjectID(ctx context.Context, projectID string) (int, error)
}
