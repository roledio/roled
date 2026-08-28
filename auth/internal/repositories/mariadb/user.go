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

type userRepository struct {
	qx interfaces.QueryExecutor
}

func NewUserRepository(qx interfaces.QueryExecutor) interfaces.UserRepository {
	return &userRepository{qx: qx}
}

func (r *userRepository) Count(ctx context.Context, req *models.GetUsersRequest) (int, error) {
	sb := sq.Select("COUNT(*)").
		From("users u").
		LeftJoin("user_roles ur ON u.id = ur.user_id").
		LeftJoin("roles r ON ur.role_id = r.id")
	sb = r.buildUserQuery(sb, req)
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

func (r *userRepository) FindAll(ctx context.Context, req *models.GetUsersRequest) ([]interfaces.UserAndRole, error) {
	sb := sq.Select("u.*", "COALESCE(r.id, '') as role_id", "COALESCE(r.name, '') as role_name").
		From("users u").
		LeftJoin("user_roles ur ON u.id = ur.user_id").
		LeftJoin("roles r ON ur.role_id = r.id")
	sb = r.buildUserQuery(sb, req)
	// Apply pagination and sorting
	req.SetDefaults()
	sb = sb.Offset(uint64(req.Offset())).Limit(uint64(req.Limit()))
	sortBy := strings.TrimSpace(req.SortBy)
	sortDir := repositories.FixSortDir(req.SortDir)
	mapSortBy := map[string]string{
		"display_name":     "u.display_name",
		"email":            "u.email",
		"is_active":        "u.is_active",
		"external_user_id": "u.external_user_id",
		"created_at":       "u.created_at",
		"updated_at":       "u.updated_at",
		"role_name":        "r.name",
	}
	if column, ok := mapSortBy[sortBy]; ok {
		sortBy = column
	} else {
		sortBy = "u.created_at" // Default sorting is by created_at descending
		sortDir = "DESC"
	}
	sb = sb.OrderBy(sortBy + " " + sortDir)
	// Build and execute query
	q, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var users []interfaces.UserAndRole
	err = r.qx.SelectContext(ctx, &users, q, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) buildUserQuery(sb sq.SelectBuilder, req *models.GetUsersRequest) sq.SelectBuilder {
	sb = sb.Where(sq.Eq{"u.project_id": req.ProjectID, "u.deleted_at": nil})
	search := strings.TrimSpace(req.Search)
	if search != "" {
		like := repositories.WithPercentAround(search)
		sb = sb.Where(sq.Or{
			sq.Like{"u.email": like},
			sq.Like{"u.display_name": like},
			sq.Like{"u.external_user_id": like},
		})
	}
	if req.IsActive != nil {
		sb = sb.Where(sq.Eq{"u.is_active": *req.IsActive})
	}
	if req.RoleID != nil {
		sb = sb.Where(sq.Eq{"r.id": *req.RoleID})
	}
	return sb
}

func (r *userRepository) FindByProjectIDAndEmail(ctx context.Context, projectID, email string) (*entities.User, error) {
	var q = "SELECT * FROM users WHERE project_id = ? AND email = ? AND deleted_at IS NULL LIMIT 1"
	var user entities.User
	err := r.qx.GetContext(ctx, &user, q, projectID, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByProjectIDAndExternalUserID(ctx context.Context, projectID, externalUserID string) (*entities.User, error) {
	var q = "SELECT * FROM users WHERE project_id = ? AND external_user_id = ? AND deleted_at IS NULL LIMIT 1"
	var user entities.User
	err := r.qx.GetContext(ctx, &user, q, projectID, externalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByProjectIDAndExternalUserIDJoinRole(ctx context.Context, projectID, externalUserID string) (*interfaces.UserAndRole, error) {
	var q = `SELECT u.*, COALESCE(r.name, '') as role_name, COALESCE(r.id, '') as role_id
			FROM users u
			LEFT JOIN user_roles ur ON u.id = ur.user_id
			LEFT JOIN roles r ON ur.role_id = r.id
			WHERE u.project_id = ? AND u.external_user_id = ? AND u.deleted_at IS NULL LIMIT 1`
	var userRole interfaces.UserAndRole
	err := r.qx.GetContext(ctx, &userRole, q, projectID, externalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &userRole, err
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	var q = `INSERT INTO users (
				id,
				account_id,
				project_id,
				email,
				password_hash,
				external_user_id,	
				display_name,
				avatar_url,
				is_active
			) VALUES (
				:id,
				:account_id,
				:project_id,
				:email,
				:password_hash,
				:external_user_id,
				:display_name,
				:avatar_url,
				:is_active
			)`
	_, err := r.qx.NamedExecContext(ctx, q, user)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	var q = "SELECT * FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1"
	var user entities.User
	err := r.qx.GetContext(ctx, &user, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByIDAndProjectID(ctx context.Context, id string, projectID string) (*entities.User, error) {
	var q = `SELECT * FROM users WHERE id = ? AND project_id = ? AND deleted_at IS NULL LIMIT 1`
	var user entities.User
	err := r.qx.GetContext(ctx, &user, q, id, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByIDAndProjectIDJoinRole(ctx context.Context, userID, projectID string) (*interfaces.UserAndRole, error) {
	var q = `SELECT u.*, COALESCE(r.name, '') as role_name, COALESCE(r.id, '') as role_id
			FROM users u
			LEFT JOIN user_roles ur ON u.id = ur.user_id
			LEFT JOIN roles r ON ur.role_id = r.id
			WHERE u.id = ? AND u.project_id = ? AND u.deleted_at IS NULL LIMIT 1`
	var userRole interfaces.UserAndRole
	err := r.qx.GetContext(ctx, &userRole, q, userID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &userRole, err
}

func (r *userRepository) SetEmailVerified(ctx context.Context, userID string) (int, error) {
	var q = `UPDATE users SET 
				email_verified_at = NOW(4), 
				updated_at = NOW(4) 
			WHERE id = ? 
			AND deleted_at IS NULL 
			AND email_verified_at IS NULL`
	return execOne(ctx, r.qx, q, userID)
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) (int, error) {
	var q = "UPDATE users SET password_hash = ?, updated_at = NOW(4) WHERE id = ? AND deleted_at IS NULL"
	return execOne(ctx, r.qx, q, passwordHash, userID)
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) (int, error) {
	var q = `UPDATE users SET
				account_id = :account_id,
				project_id = :project_id,
				email = :email,
				email_verified_at = :email_verified_at,
				password_hash = :password_hash,
				external_user_id = :external_user_id,
				display_name = :display_name,
				is_active = :is_active,
				avatar_url = :avatar_url,
				updated_at = NOW(4)
			WHERE id = :id AND deleted_at IS NULL`
	return namedExecOne(ctx, r.qx, q, user)
}

func (r *userRepository) DeleteByID(ctx context.Context, userID string) (int, error) {
	var q = `UPDATE users SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, userID)
}

func (r *userRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	var q = `UPDATE users SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE account_id = ? AND deleted_at IS NULL`
	return exec(ctx, r.qx, q, accountID)
}

func (r *userRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	var q = `UPDATE users SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE project_id = ? AND deleted_at IS NULL`
	return exec(ctx, r.qx, q, projectID)
}
