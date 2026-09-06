package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type OAuthConnectionRepository interface {
	FindByProjectID(ctx context.Context, projectID string) ([]entities.OAuthConnection, error)
}
