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

type resourceRepository struct {
	qx interfaces.QueryExecutor
}

func NewResourceRepository(qx interfaces.QueryExecutor) interfaces.ResourceRepository {
	return &resourceRepository{qx: qx}
}

func (r *resourceRepository) FindByProjectID(ctx context.Context, projectID string) ([]entities.Resource, error) {
	var q = `SELECT * FROM resources WHERE project_id = ?`
	var results []entities.Resource
	err := r.qx.SelectContext(ctx, &results, q, projectID)
	return results, err
}

func (r *resourceRepository) Create(ctx context.Context, resources []entities.Resource) (int, error) {
	if len(resources) == 0 {
		return 0, nil
	}

	builder := sq.
		Insert("resources").
		Columns("id", "account_id", "project_id", "name", "code", "description", "is_default")

	for _, resource := range resources {
		builder = builder.Values(
			resource.ID,
			resource.AccountID,
			resource.ProjectID,
			resource.Name,
			resource.Code,
			resource.Description,
			resource.IsDefault,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	return exec(ctx, r.qx, query, args...)
}

func (r *resourceRepository) Count(ctx context.Context, req *models.GetResourcesRequest) (int, error) {
	sb := sq.
		Select("COUNT(*)").
		From("resources")
	sb = r.buildResourceQuery(sb, req)
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

func (r *resourceRepository) FindAll(ctx context.Context, req *models.GetResourcesRequest) ([]entities.Resource, error) {
	sb := sq.
		Select("resources.*").
		From("resources")
	sb = r.buildResourceQuery(sb, req)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name": "resources.name",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "resources.name" // Default sorting is by resource name ascending
		sortDir = "ASC"
	}
	sb = sb.OrderBy("resources.is_default ASC", sortBy+" "+sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var resources []entities.Resource
	err = r.qx.SelectContext(ctx, &resources, q, args...)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *resourceRepository) buildResourceQuery(sb sq.SelectBuilder, req *models.GetResourcesRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{"resources.project_id": req.ProjectID})
	search := strings.TrimSpace(req.Search)
	if search != "" {
		like := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"resources.name": like},
			sq.Like{"resources.code": like},
			sq.Like{"resources.description": like},
			sq.Expr(`
			EXISTS (
				SELECT 1
				FROM permissions 
				WHERE permissions.resource_id = resources.id
				AND (permissions.name LIKE ? OR permissions.code LIKE ? OR permissions.description LIKE ?)
			)`, like, like, like),
		})
	}
	if req.IsDefault != nil {
		sb = sb.Where(sq.Eq{"resources.is_default": *req.IsDefault})
	}
	return sb
}

func (r *resourceRepository) FindByIDAndProjectID(ctx context.Context, resourceID string, projectID string) (*entities.Resource, error) {
	var q = `SELECT * FROM resources WHERE id = ? AND project_id = ?`
	var resource entities.Resource
	err := r.qx.GetContext(ctx, &resource, q, resourceID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &resource, err
}

func (r *resourceRepository) FindByProjectIDAndCode(ctx context.Context, projectID string, code string) (*entities.Resource, error) {
	var q = `SELECT * FROM resources WHERE project_id = ? AND code = ?`
	var resource entities.Resource
	err := r.qx.GetContext(ctx, &resource, q, projectID, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &resource, err
}

func (r *resourceRepository) Update(ctx context.Context, resource *entities.Resource) (int, error) {
	q := `UPDATE resources SET 
			name = :name, 
			code = :code,
			description = :description,
			updated_at = NOW(4) 
		WHERE id = :id AND project_id = :project_id`
	return namedExecOne(ctx, r.qx, q, resource)
}

func (r *resourceRepository) Delete(ctx context.Context, resource *entities.Resource) (int, error) {
	q := `DELETE FROM resources WHERE id = ?`
	return execOne(ctx, r.qx, q, resource.ID)
}
