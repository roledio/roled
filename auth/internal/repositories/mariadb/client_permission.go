package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type clientPermissionRepository struct {
	qx interfaces.QueryExecutor
}

func NewClientPermissionRepository(qx interfaces.QueryExecutor) interfaces.ClientPermissionRepository {
	return &clientPermissionRepository{qx: qx}
}

func (r *clientPermissionRepository) Create(ctx context.Context, clientPermissions []entities.ClientPermission) error {
	if len(clientPermissions) == 0 {
		return nil
	}

	builder := sq.
		Insert("client_permissions").
		Columns("client_id", "permission_id")

	for _, permission := range clientPermissions {
		builder = builder.Values(
			permission.ClientID,
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

func (r *clientPermissionRepository) DeleteByClientID(ctx context.Context, clientID string) (int, error) {
	q := `DELETE FROM client_permissions WHERE client_id = ?`
	return exec(ctx, r.qx, q, clientID)
}

func (r *clientPermissionRepository) FindByClientID(ctx context.Context, clientID string) ([]entities.ClientPermission, error) {
	var q = `SELECT * FROM client_permissions WHERE client_id = ?`
	var results []entities.ClientPermission
	err := r.qx.SelectContext(ctx, &results, q, clientID)
	return results, err
}

func (r *clientPermissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]entities.RolePermission, error) {
	var q = `SELECT * FROM role_permissions WHERE role_id = ?`
	var results []entities.RolePermission
	err := r.qx.SelectContext(ctx, &results, q, roleID)
	return results, err
}
