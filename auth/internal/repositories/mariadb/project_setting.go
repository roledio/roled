package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/repositories/interfaces"
)

type projectSettingRepository struct {
	qx interfaces.QueryExecutor
}

func NewProjectSettingRepository(qx interfaces.QueryExecutor) interfaces.ProjectSettingRepository {
	return &projectSettingRepository{qx: qx}
}

func (r *projectSettingRepository) FindByProjectID(ctx context.Context, projectID string) (*entities.ProjectSetting, error) {
	var q = "SELECT * FROM project_settings WHERE project_id = ? LIMIT 1"
	var projectSetting entities.ProjectSetting
	err := r.qx.GetContext(ctx, &projectSetting, q, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &projectSetting, err
}

func (r *projectSettingRepository) Create(ctx context.Context, projectSetting *entities.ProjectSetting) error {
	var q = `INSERT INTO project_settings (
			id,
			project_id, 
			is_signup_enabled, 
			default_signup_role_id,
			is_signup_verify_email,
			is_forgot_password_enabled,
			is_allow_temp_email
		) VALUES (
			:id,
			:project_id, 
			:is_signup_enabled, 
			:default_signup_role_id,
			:is_signup_verify_email,
			:is_forgot_password_enabled,
			:is_allow_temp_email
		)`
	_, err := r.qx.NamedExecContext(ctx, q, projectSetting)
	return err
}

func (r *projectSettingRepository) Update(ctx context.Context, projectSetting *entities.ProjectSetting) (int, error) {
	var q = `UPDATE project_settings SET
			is_signup_enabled = :is_signup_enabled,
			default_signup_role_id = :default_signup_role_id,
			is_signup_verify_email = :is_signup_verify_email,
			is_forgot_password_enabled = :is_forgot_password_enabled,
			is_allow_temp_email = :is_allow_temp_email,
			updated_at = NOW(4)
		WHERE project_id = :project_id`
	return namedExecOne(ctx, r.qx, q, projectSetting)
}
