package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type BrandingRepository interface {
	FindByProjectID(context.Context, string) (*entities.Branding, error)
	Upsert(context.Context, *entities.Branding) (int, error)
}
