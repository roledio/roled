package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	"github.com/roledio/roled/auth/pkg/repositories"
)

type roleRepository struct {
	qx interfaces.QueryExecutor
}

func NewRoleRepository(qx interfaces.QueryExecutor) interfaces.RoleRepository {
	return &roleRepository{qx: qx}
}

func (r *roleRepository) Count(ctx context.Context, req *models.GetProjectRolesRequest) (int, error) {
	sb := sq.Select("COUNT(*)").From("roles")
	sb = r.buildRoleQuery(sb, req)
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

func (r *roleRepository) FindAll(ctx context.Context, req *models.GetProjectRolesRequest) ([]entities.Role, error) {
	sb := sq.Select("*").From("roles")
	sb = r.buildRoleQuery(sb, req)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name":       "name",
		"code":       "code",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "created_at" // Default sorting is by created_at descending
		sortDir = "DESC"
	}
	sb = sb.OrderBy(sortBy + " " + sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var roles []entities.Role
	err = r.qx.SelectContext(ctx, &roles, q, args...)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) buildRoleQuery(sb sq.SelectBuilder, req *models.GetProjectRolesRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{"project_id": req.ProjectID, "deleted_at": nil})
	search := strings.TrimSpace(req.Search)
	if search != "" {
		like := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"code": like},
			sq.Like{"name": like},
			sq.Like{"description": like},
		})
	}
	return sb
}

func (r *roleRepository) FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Role, error) {
	var q = "SELECT * FROM roles WHERE project_id = ? AND deleted_at IS NULL AND code = ? LIMIT 1"
	var role entities.Role
	err := r.qx.GetContext(ctx, &role, q, projectID, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) Create(ctx context.Context, role *entities.Role) error {
	q := `INSERT INTO roles (
			id, 
			account_id,
			project_id, 
			code, 
			name, 
			description
		) VALUES (
		 	:id, 
			:account_id,
			:project_id, 
			:code, 
			:name, 
			:description
		)`
	_, err := namedExecOne(ctx, r.qx, q, role)
	return err
}

func (r *roleRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Role, error) {
	var q = "SELECT * FROM roles WHERE id = ? AND project_id = ? AND deleted_at IS NULL LIMIT 1"
	var role entities.Role
	err := r.qx.GetContext(ctx, &role, q, id, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*entities.Role, error) {
	var q = "SELECT * FROM roles WHERE id = ? AND deleted_at IS NULL LIMIT 1"
	var role entities.Role
	err := r.qx.GetContext(ctx, &role, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) Update(ctx context.Context, role *entities.Role) (int, error) {
	q := `UPDATE roles SET
			code = :code,
			name = :name,
			description = :description,
			updated_at = NOW(4)
		WHERE id = :id AND deleted_at IS NULL`
	return namedExecOne(ctx, r.qx, q, role)
}

func (r *roleRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	q := `UPDATE roles SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, id)
}

func (r *roleRepository) FindByUserID(ctx context.Context, userID string) (*entities.Role, error) {
	var q = `SELECT r.* FROM roles r 
				JOIN user_roles ur ON r.id = ur.role_id 
				WHERE ur.user_id = ? AND r.deleted_at IS NULL LIMIT 1`
	var role entities.Role
	err := r.qx.GetContext(ctx, &role, q, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &role, err
}
