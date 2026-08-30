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

type accountRepository struct {
	qx interfaces.QueryExecutor
}

func NewAccountRepository(qx interfaces.QueryExecutor) interfaces.AccountRepository {
	return &accountRepository{qx: qx}
}

func (r *accountRepository) FindByID(ctx context.Context, id string) (*entities.Account, error) {
	var q = "SELECT * FROM accounts WHERE id = ? AND deleted_at IS NULL LIMIT 1"
	var account entities.Account
	err := r.qx.GetContext(ctx, &account, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &account, err
}

func (r *accountRepository) Create(ctx context.Context, account *entities.Account) error {
	var q = `INSERT INTO accounts (
				id,
				name,
				description,
				is_active,
				is_system
			) VALUES (
				:id,
				:name,
				:description,
				:is_active,
				:is_system
			)`
	_, err := r.qx.NamedExecContext(ctx, q, account)
	return err
}

func (r *accountRepository) FindAll(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) ([]entities.Account, error) {
	sb := sq.Select("*").From("accounts")
	sb = r.buildAccountQuery(sb, req, filterAccountID)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	if req.SortBy == "" { // If not specified, use default: updated_at DESC
		req.SortBy = "updated_at"
		req.SortDir = "DESC"
	}
	sb = sb.OrderBy(req.SortBy + " " + req.SortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var accounts []entities.Account
	err = r.qx.SelectContext(ctx, &accounts, q, args...)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepository) Count(ctx context.Context, req *models.GetAccountsRequest, filterAccountID *string) (int, error) {
	sb := sq.Select("COUNT(*)").From("accounts")
	sb = r.buildAccountQuery(sb, req, filterAccountID)
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

func (r *accountRepository) buildAccountQuery(sb sq.SelectBuilder, req *models.GetAccountsRequest, filterAccountID *string) sq.SelectBuilder {
	if filterAccountID != nil {
		sb = sb.Where(sq.Eq{"id": *filterAccountID, "is_system": false})
	}
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

func (r *accountRepository) Update(ctx context.Context, account *entities.Account) (int, error) {
	var q = `UPDATE accounts SET
				name = :name,
				description = :description,
				is_active = :is_active,
				updated_at = :updated_at
			WHERE id = :id AND deleted_at IS NULL`
	return namedExecOne(ctx, r.qx, q, account)
}

func (r *accountRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	var q = `UPDATE accounts SET 
				deleted_at = NOW(4), 
				updated_at = NOW(4) 
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, id)
}
