package mariadb

import (
	"context"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type userRoleRepository struct {
	qx interfaces.QueryExecutor
}

func NewUserRoleRepository(qx interfaces.QueryExecutor) interfaces.UserRoleRepository {
	return &userRoleRepository{qx: qx}
}

func (r *userRoleRepository) Create(ctx context.Context, userRole *entities.UserRole) error {
	var q = `INSERT INTO user_roles (
				user_id,
				role_id
			) VALUES (
				:user_id,
				:role_id
			)`
	_, err := r.qx.NamedExecContext(ctx, q, userRole)
	return err
}

func (r *userRoleRepository) DeleteByUserID(ctx context.Context, userID string) (int, error) {
	var q = `DELETE FROM user_roles WHERE user_id = ?`
	return exec(ctx, r.qx, q, userID)
}

func (r *userRoleRepository) FindUserIDsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	var q = `SELECT user_id FROM user_roles WHERE role_id = ?`
	var userIDs []string
	err := r.qx.SelectContext(ctx, &userIDs, q, roleID)
	return userIDs, err
}
