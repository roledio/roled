package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/internal/repositories/interfaces"
	"github.com/roledio/roled/pkg/repositories"
)

type clientRepository struct {
	qx interfaces.QueryExecutor
}

func NewClientRepository(qx interfaces.QueryExecutor) interfaces.ClientRepository {
	return &clientRepository{qx: qx}
}

func (r *clientRepository) FindByID(ctx context.Context, id string) (*entities.Client, error) {
	var q = "SELECT * FROM clients WHERE id = ? AND deleted_at IS NULL LIMIT 1"
	var client entities.Client
	err := r.qx.GetContext(ctx, &client, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &client, err
}

func (r *clientRepository) FindByProjectIDAndIsDefault(ctx context.Context, projectID string, isDefault bool) (*entities.Client, error) {
	var q = "SELECT * FROM clients WHERE project_id = ? AND is_default = ? AND deleted_at IS NULL LIMIT 1"
	var client entities.Client
	err := r.qx.GetContext(ctx, &client, q, projectID, isDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &client, err
}

func (r *clientRepository) Create(ctx context.Context, client *entities.Client) error {
	q := `INSERT INTO clients (
			id, 
			account_id, 
			project_id, 
			name, 
			description,
			secret_encrypted, 
			is_active, 
			is_default
		) VALUES (
		 	:id, 
			:account_id, 
			:project_id, 
			:name, 
			:description,
			:secret_encrypted, 
			:is_active, 
			:is_default
		)`
	_, err := namedExecOne(ctx, r.qx, q, client)
	return err
}

func (r *clientRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	q := `UPDATE clients SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE project_id = ? AND deleted_at IS NULL`
	return exec(ctx, r.qx, q, projectID)
}

func (r *clientRepository) Count(ctx context.Context, req *models.GetClientsRequest) (int, error) {
	sb := sq.Select("COUNT(*)").From("clients")
	sb = r.buildClientQuery(sb, req)
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

func (r *clientRepository) FindAll(ctx context.Context, req *models.GetClientsRequest) ([]entities.Client, error) {
	sb := sq.Select("*").From("clients")
	sb = r.buildClientQuery(sb, req)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name":       "name",
		"is_active":  "is_active",
		"created_at": "created_at",
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
	var clients []entities.Client
	err = r.qx.SelectContext(ctx, &clients, q, args...)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *clientRepository) buildClientQuery(sb sq.SelectBuilder, req *models.GetClientsRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{"project_id": req.ProjectID, "deleted_at": nil})
	search := strings.TrimSpace(req.Search)
	if req.Search != "" {
		searchPattern := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"name": searchPattern},
			sq.Like{"description": searchPattern},
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
	return sb
}

func (r *clientRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.Client, error) {
	var q = "SELECT * FROM clients WHERE id = ? AND project_id = ? AND deleted_at IS NULL LIMIT 1"
	var client entities.Client
	err := r.qx.GetContext(ctx, &client, q, id, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &client, err
}

func (r *clientRepository) Update(ctx context.Context, client *entities.Client) (int, error) {
	q := `UPDATE clients SET
			name = :name,
			description = :description,
			is_active = :is_active,
			updated_at = NOW(4)
		WHERE id = :id AND deleted_at IS NULL`
	return namedExecOne(ctx, r.qx, q, client)
}

func (r *clientRepository) Delete(ctx context.Context, client *entities.Client) (int, error) {
	q := `UPDATE clients SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, client.ID)
}
