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

type projectRepository struct {
	qx interfaces.QueryExecutor
}

func NewProjectRepository(qx interfaces.QueryExecutor) interfaces.ProjectRepository {
	return &projectRepository{qx: qx}
}

func (r *projectRepository) FindByID(ctx context.Context, id string) (*entities.Project, error) {
	var q = "SELECT * FROM projects WHERE id = ? AND deleted_at IS NULL LIMIT 1"
	var project entities.Project
	err := r.qx.GetContext(ctx, &project, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &project, err
}

func (r *projectRepository) FindByIDAndAccountID(ctx context.Context, id string, accountID string) (*entities.Project, error) {
	var q = "SELECT * FROM projects WHERE id = ? AND account_id = ? AND deleted_at IS NULL LIMIT 1"
	var project entities.Project
	err := r.qx.GetContext(ctx, &project, q, id, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &project, err
}

func (r *projectRepository) FindSystem(ctx context.Context) (*entities.Project, error) {
	var q = "SELECT * FROM projects WHERE is_system = true AND deleted_at IS NULL LIMIT 1"
	var project entities.Project
	err := r.qx.GetContext(ctx, &project, q)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &project, err
}

func (r *projectRepository) Count(ctx context.Context, req *models.GetProjectsRequest, accountID string) (int, error) {
	sb := sq.Select("COUNT(*)").From("projects")
	sb = r.buildProjectQuery(sb, req, accountID)
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

func (r *projectRepository) FindAll(ctx context.Context, req *models.GetProjectsRequest, accountID string) ([]entities.Project, error) {
	sb := sq.Select("*").From("projects")
	sb = r.buildProjectQuery(sb, req, accountID)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name":       "name",
		"created_at": "created_at",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "created_at"
		sortDir = "DESC"
	}
	sb = sb.OrderBy(sortBy + " " + sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var projects []entities.Project
	err = r.qx.SelectContext(ctx, &projects, q, args...)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) buildProjectQuery(sb sq.SelectBuilder, req *models.GetProjectsRequest, accountID string) sq.SelectBuilder {
	// Set the order to match index ordering: account_id first, then is_system
	sb = sb.Where(sq.Eq{"account_id": accountID})
	search := strings.TrimSpace(req.Search)
	if search != "" {
		likePattern := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"name": likePattern},
			sq.Like{"description": likePattern},
		})
	}
	if req.IsActive != nil {
		sb = sb.Where(sq.Eq{"is_active": *req.IsActive})
	}
	if req.CreatedAtSince != nil {
		sb = sb.Where(sq.GtOrEq{"created_at": *req.CreatedAtSince})
	}
	if req.CreatedAtUntil != nil {
		sb = sb.Where(sq.LtOrEq{"created_at": *req.CreatedAtUntil})
	}
	sb = sb.Where(sq.Eq{"deleted_at": nil})
	return sb
}

func (r *projectRepository) Create(ctx context.Context, project *entities.Project) error {
	q := `INSERT INTO projects (
			id, 
			account_id, 
			name, 
			description, 
			logo_url, 
			is_active, 
			is_system
		) VALUES (
			:id, 
			:account_id,
			:name, 
			:description, 
			:logo_url, 
			:is_active, 
			:is_system
		)`
	_, err := namedExecOne(ctx, r.qx, q, project)
	return err
}

func (r *projectRepository) Update(ctx context.Context, project *entities.Project) (int, error) {
	q := `UPDATE projects SET
			name = :name,
			description = :description,
			logo_url = :logo_url,
			is_active = :is_active,
			updated_at = NOW(4)
		WHERE id = :id AND deleted_at IS NULL`
	return namedExecOne(ctx, r.qx, q, project)
}

func (r *projectRepository) Delete(ctx context.Context, project *entities.Project) (int, error) {
	var q = `UPDATE projects SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, project.ID)
}
