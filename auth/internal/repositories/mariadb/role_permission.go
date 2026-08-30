package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type rolePermissionRepository struct {
	qx interfaces.QueryExecutor
}

func NewRolePermissionRepository(qx interfaces.QueryExecutor) interfaces.RolePermissionRepository {
	return &rolePermissionRepository{qx: qx}
}

func (r *rolePermissionRepository) Create(ctx context.Context, rolePermissions []entities.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}

	builder := sq.
		Insert("role_permissions").
		Columns("role_id", "permission_id")

	for _, permission := range rolePermissions {
		builder = builder.Values(
			permission.RoleID,
			permission.PermissionID,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = exec(ctx, r.qx, query, args...)
	return err
}

func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID string) (int, error) {
	q := `DELETE FROM role_permissions WHERE role_id = ?`
	return exec(ctx, r.qx, q, roleID)
}
