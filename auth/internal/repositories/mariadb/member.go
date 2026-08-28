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

type memberRepository struct {
	qx interfaces.QueryExecutor
}

func NewMemberRepository(qx interfaces.QueryExecutor) interfaces.MemberRepository {
	return &memberRepository{qx: qx}
}

func (r *memberRepository) Create(ctx context.Context, member *entities.Member) error {
	var q = `INSERT INTO members (
				id,
				account_id,
				user_id,
				is_admin
			) VALUES (
				:id,
				:account_id,
				:user_id,
				:is_admin
			)`
	_, err := r.qx.NamedExecContext(ctx, q, member)
	return err
}

func (r *memberRepository) FindByAccountIDAndUserID(ctx context.Context, accountID string, userID string) (*entities.Member, error) {
	var q = "SELECT * FROM members WHERE account_id = ? AND deleted_at IS NULL AND user_id = ?"
	var member entities.Member
	err := r.qx.GetContext(ctx, &member, q, accountID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &member, err
}

func (r *memberRepository) FindAll(ctx context.Context, req *models.GetMembersRequest) ([]interfaces.MemberUser, error) {
	sb := sq.Select(
		"members.*",
		"users.email",
		"users.display_name",
		"users.avatar_url",
		"users.is_active",
		"users.email_verified_at IS NOT NULL AS is_verified").
		From("members").
		Join("users ON users.id = members.user_id")

	sb = r.buildMemberQuery(sb, req)

	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"name":       "users.display_name",
		"is_active":  "users.is_active",
		"is_admin":   "members.is_admin",
		"updated_at": "members.updated_at",
		"created_at": "members.created_at",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "members.created_at"
		sortDir = "DESC"
	}
	sb = sb.OrderBy(sortBy + " " + sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var members []interfaces.MemberUser
	err = r.qx.SelectContext(ctx, &members, q, args...)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *memberRepository) Count(ctx context.Context, req *models.GetMembersRequest) (int, error) {
	sb := sq.Select("COUNT(*)").
		From("members").
		Join("users ON users.id = members.user_id")
	sb = r.buildMemberQuery(sb, req)
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

func (r *memberRepository) buildMemberQuery(sb sq.SelectBuilder, req *models.GetMembersRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{
		"members.account_id": req.AccountID, // Account ID is required, either from the JWT or the request (for system account)
		"members.deleted_at": nil},
	)
	search := strings.TrimSpace(req.Search)
	if search != "" {
		likePattern := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"users.email": likePattern},
			sq.Like{"users.display_name": likePattern},
		})
	}
	if req.IsActive != nil {
		sb = sb.Where(sq.Eq{"users.is_active": *req.IsActive})
	}
	if req.IsVerified != nil {
		if *req.IsVerified {
			sb = sb.Where(sq.NotEq{"users.email_verified_at": nil})
		} else {
			sb = sb.Where(sq.Eq{"users.email_verified_at": nil})
		}
	}
	if req.CreatedAtSince != nil {
		sb = sb.Where(sq.GtOrEq{"members.created_at": *req.CreatedAtSince})
	}
	if req.CreatedAtUntil != nil {
		sb = sb.Where(sq.LtOrEq{"members.created_at": *req.CreatedAtUntil})
	}
	if req.IsAdmin != nil {
		sb = sb.Where(sq.Eq{"members.is_admin": *req.IsAdmin})
	}
	return sb
}

func (r *memberRepository) FindByID(ctx context.Context, id string) (*entities.Member, error) {
	var q = "SELECT * FROM members WHERE id = ? AND deleted_at IS NULL"
	var member entities.Member
	err := r.qx.GetContext(ctx, &member, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &member, err
}

func (r *memberRepository) Delete(ctx context.Context, member *entities.Member) (int, error) {
	var q = `UPDATE members SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, member.ID)
}

func (r *memberRepository) Update(ctx context.Context, member *entities.Member) (int, error) {
	var q = `UPDATE members SET
				is_admin = ?,
				updated_at = NOW(4)
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, member.IsAdmin, member.ID)
}

func (r *memberRepository) CountByAccountID(ctx context.Context, accountID string, isAdmin *bool) (int, error) {
	sb := sq.Select("COUNT(*)").From("members").Where(sq.Eq{"account_id": accountID, "deleted_at": nil})
	if isAdmin != nil {
		sb = sb.Where(sq.Eq{"is_admin": *isAdmin})
	}
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

func (r *memberRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	var q = `UPDATE members SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE account_id = ? AND deleted_at IS NULL`
	return exec(ctx, r.qx, q, accountID)
}

func (r *memberRepository) FindByIDJoinUser(ctx context.Context, id string) (*interfaces.MemberUser, error) {
	var q = `
		SELECT 
			members.*, 
			users.email, 
			users.display_name, 
			users.is_active,
			users.avatar_url,
			users.email_verified_at IS NOT NULL AS is_verified
		FROM members 
		INNER JOIN users ON users.id = members.user_id 
		WHERE members.id = ? AND members.deleted_at IS NULL`
	var member interfaces.MemberUser
	err := r.qx.GetContext(ctx, &member, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &member, err
}
