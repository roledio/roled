package mariadb

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/roledio/roled/pkg/repositories"
)

type permissionRepository struct {
	qx interfaces.QueryExecutor
}

func NewPermissionRepository(qx interfaces.QueryExecutor) interfaces.PermissionRepository {
	return &permissionRepository{qx: qx}
}

func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]interfaces.PermissionResource, error) {
	var q = `SELECT p.*, r.name as resource_name, r.code as resource_code
		FROM permissions p
		INNER JOIN resources r ON r.id = p.resource_id
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ?
		ORDER BY r.name, p.name
	`
	var results []interfaces.PermissionResource
	err := r.qx.SelectContext(ctx, &results, q, roleID)
	return results, err
}

func (r *permissionRepository) FindByClientID(ctx context.Context, clientID string) ([]interfaces.PermissionResource, error) {
	var q = `SELECT p.*, r.name as resource_name, r.code as resource_code
		FROM permissions p
		INNER JOIN resources r ON r.id = p.resource_id
		INNER JOIN client_permissions cp ON cp.permission_id = p.id
		WHERE cp.client_id = ?
		ORDER BY r.name, p.name
	`
	var results []interfaces.PermissionResource
	err := r.qx.SelectContext(ctx, &results, q, clientID)
	return results, err
}

func (r *permissionRepository) FindByResourceIDsAndSearch(ctx context.Context, resourceIDs []string, search string) ([]entities.Permission, error) {
	if len(resourceIDs) == 0 {
		return []entities.Permission{}, nil
	}
	q, args, err := sqlx.In(`SELECT * FROM permissions WHERE resource_id IN (?)`, resourceIDs)
	if err != nil {
		return nil, err
	}
	q = r.qx.Rebind(q)
	if search = strings.TrimSpace(search); search != "" {
		q += ` AND (name LIKE ? OR code LIKE ? OR description LIKE ?)`
		like := repositories.WithPercentAround(search)
		args = append(args, like, like, like)
	}
	q += ` ORDER BY name`
	var results []entities.Permission
	err = r.qx.SelectContext(ctx, &results, q, args...)
	return results, err
}

func (r *permissionRepository) Create(ctx context.Context, permissions []entities.Permission) (int, error) {
	if len(permissions) == 0 {
		return 0, nil
	}

	builder := sq.
		Insert("permissions").
		Columns("id", "resource_id", "name", "code", "description", "is_default")

	for _, permission := range permissions {
		builder = builder.Values(
			permission.ID,
			permission.ResourceID,
			permission.Name,
			permission.Code,
			permission.Description,
			permission.IsDefault,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	return exec(ctx, r.qx, query, args...)
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []string) ([]interfaces.PermissionResource, error) {
	if len(ids) == 0 {
		return []interfaces.PermissionResource{}, nil
	}
	q, args, err := sqlx.In(`
		SELECT p.*, r.name as resource_name, r.code as resource_code
		FROM permissions p
		INNER JOIN resources r ON r.id = p.resource_id
		WHERE p.id IN (?)
		ORDER BY r.name, p.name`, ids)
	if err != nil {
		return nil, err
	}
	q = r.qx.Rebind(q)
	var results []interfaces.PermissionResource
	err = r.qx.SelectContext(ctx, &results, q, args...)
	return results, err
}

func (r *permissionRepository) DeleteByResourceID(ctx context.Context, resourceID string) (int, error) {
	q := `DELETE FROM permissions WHERE resource_id = ?`
	return exec(ctx, r.qx, q, resourceID)
}

func (r *permissionRepository) FindAll(ctx context.Context, req *models.GetPermissionsRequest) ([]interfaces.PermissionResource, error) {
	sb := sq.
		Select("p.*, r.name as resource_name, r.code as resource_code").
		From("permissions p").
		InnerJoin("resources r ON r.id = p.resource_id")
	sb = r.buildPermissionQuery(sb, req)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name":          "p.name",
		"code":          "p.code",
		"created_at":    "p.created_at",
		"updated_at":    "p.updated_at",
		"resource_name": "r.name",
		"resource_code": "r.code",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "r.name" // Default sorting is by resource name ascending
		sortDir = "ASC"
	}
	sb = sb.OrderBy(sortBy + " " + sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var resources []interfaces.PermissionResource
	err = r.qx.SelectContext(ctx, &resources, q, args...)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *permissionRepository) Count(ctx context.Context, req *models.GetPermissionsRequest) (int, error) {
	sb := sq.
		Select("COUNT(*)").
		From("permissions p").
		InnerJoin("resources r ON r.id = p.resource_id")
	sb = r.buildPermissionQuery(sb, req)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}
	var count int
	err = r.qx.GetContext(ctx, &count, q, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *permissionRepository) buildPermissionQuery(sb sq.SelectBuilder, req *models.GetPermissionsRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{"r.project_id": req.ProjectID})
	search := strings.TrimSpace(req.Search)
	if search != "" {
		like := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"r.name": like},
			sq.Like{"p.name": like},
			sq.Like{"p.description": like},
		})
	}
	if req.IsDefault != nil {
		sb = sb.Where(sq.Eq{"p.is_default": *req.IsDefault})
	}
	return sb
}

func (r *permissionRepository) FindByProjectID(ctx context.Context, projectID string, isDefault *bool) ([]entities.Permission, error) {
	sb := sq.
		Select("p.*").
		From("permissions p").
		InnerJoin("resources r ON r.id = p.resource_id").
		Where(sq.Eq{"r.project_id": projectID})
	if isDefault != nil {
		sb = sb.Where(
			sq.Eq{"r.is_default": *isDefault},
			sq.Eq{"p.is_default": *isDefault},
		)
	}
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var permissions []entities.Permission
	err = r.qx.SelectContext(ctx, &permissions, q, args...)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
