package interfaces

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
)

type ProjectSettingRepository interface {
	FindByProjectID(ctx context.Context, projectID string) (*entities.ProjectSetting, error)
	Create(ctx context.Context, projectSetting *entities.ProjectSetting) error
	Update(ctx context.Context, projectSetting *entities.ProjectSetting) (int, error)
}
