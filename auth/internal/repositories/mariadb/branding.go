package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/pkg/repositories"
)

type brandingRepository struct{ qx repositories.QueryExecutor }

func NewBrandingRepository(qx repositories.QueryExecutor) interfaces.BrandingRepository {
	return &brandingRepository{qx: qx}
}

func (r *brandingRepository) FindByProjectID(ctx context.Context, projectID string) (*entities.Branding, error) {
	var branding entities.Branding
	err := r.qx.GetContext(ctx, &branding, "SELECT * FROM brandings WHERE project_id = ?", projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &branding, err
}

func (r *brandingRepository) Upsert(ctx context.Context, b *entities.Branding) (int, error) {
	return repositories.NamedExec(ctx, r.qx, `
	INSERT INTO brandings
	(
		project_id,
		logo_url,
		favicon_url,
		primary_color,
		rounding,
		enable_shadow,
		enable_border
	)
	VALUES
	(
		:project_id,
		:logo_url,
		:favicon_url,
		:primary_color,
		:rounding,
		:enable_shadow,
		:enable_border
	)
	ON DUPLICATE KEY UPDATE 
		logo_url=VALUES(logo_url),
		favicon_url=VALUES(favicon_url),
		primary_color=VALUES(primary_color),
		rounding=VALUES(rounding),
		enable_shadow=VALUES(enable_shadow),
		enable_border=VALUES(enable_border),
		updated_at=NOW(4)
	`, b)
}
