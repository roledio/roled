package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type accessTokenRepository struct {
	qx interfaces.QueryExecutor
}

func NewAccessTokenRepository(qx interfaces.QueryExecutor) interfaces.AccessTokenRepository {
	return &accessTokenRepository{qx: qx}
}

func (r *accessTokenRepository) Create(ctx context.Context, token *entities.AccessToken) error {
	q := `
		INSERT INTO access_tokens (
			id,
			account_id,
			project_id,
			client_id,
			user_id,
			refresh_token_id,
			auth_code_id,
			grant_type,
			issued_at,
			expires_in,
			status
		) VALUES (
			:id,
			:account_id,
			:project_id,
			:client_id,
			:user_id,
			:refresh_token_id,
			:auth_code_id,
			:grant_type,
			:issued_at,
			:expires_in,
			:status
		)
	`
	_, err := r.qx.NamedExecContext(ctx, q, token)
	return err
}

func (r *accessTokenRepository) FindByID(ctx context.Context, id string) (*entities.AccessToken, error) {
	q := "SELECT * FROM access_tokens WHERE id = ? AND deleted_at IS NULL"
	var token entities.AccessToken
	err := r.qx.GetContext(ctx, &token, q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &token, err
}

func (r *accessTokenRepository) UpdateAsIssued(ctx context.Context, token *entities.AccessToken) (int, error) {
	q := `
		UPDATE access_tokens SET
			updated_at = NOW(4),
			status = :status,
			issued_at = :issued_at,
			expires_in = :expires_in,
			refresh_token_id = :refresh_token_id
		WHERE id = :id AND deleted_at IS NULL
	`
	return namedExecOne(ctx, r.qx, q, token)
}

func (r *accessTokenRepository) UpdateAsRevoked(ctx context.Context, id string) (int, error) {
	q := `
		UPDATE access_tokens SET
			updated_at = NOW(4),
			revoked_at = NOW(4),
			status = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	return execOne(ctx, r.qx, q, constants.AccessTokenStatusRevoked, id)
}

func (r *accessTokenRepository) DeleteByAccountID(ctx context.Context, accountID string) (int, error) {
	q := `
		UPDATE access_tokens SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE account_id = ? AND deleted_at IS NULL
	`
	return exec(ctx, r.qx, q, accountID)
}

func (r *accessTokenRepository) DeleteByUserID(ctx context.Context, userID string) (int, error) {
	q := `
		UPDATE access_tokens SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE user_id = ? AND deleted_at IS NULL
	`
	return exec(ctx, r.qx, q, userID)
}

func (r *accessTokenRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	q := `
		UPDATE access_tokens SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE project_id = ? AND deleted_at IS NULL
	`
	return exec(ctx, r.qx, q, projectID)
}

func (r *accessTokenRepository) DeleteByClientID(ctx context.Context, clientID string) (int, error) {
	q := `
		UPDATE access_tokens SET
			deleted_at = NOW(4),
			updated_at = NOW(4)
		WHERE client_id = ? AND deleted_at IS NULL
	`
	return exec(ctx, r.qx, q, clientID)
}

func (r *accessTokenRepository) FindByIDJoin(ctx context.Context, id string) (*interfaces.AccessTokenJoinResult, error) {
	q := `SELECT
			at.id,
			at.issued_at,
			at.expires_in,

			p.id   AS project_id,
			p.name AS project_name,
			p.description AS project_description,
			p.logo_url AS project_logo_url,

			c.id   AS client_id,
			c.name AS client_name,

			u.id   AS user_id,
			u.display_name AS user_display_name,
			u.email AS user_email,
			u.external_user_id AS user_external_user_id,
			u.avatar_url AS user_avatar_url,

			r.id   AS role_id,
			r.code AS role_code,
			r.name AS role_name,
			r.description AS role_description

		FROM access_tokens at
		JOIN projects p ON p.id = at.project_id
		JOIN clients c  ON c.id = at.client_id
		LEFT JOIN users u ON u.id = at.user_id
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE at.id = ?
		AND at.status = ?
		AND at.deleted_at IS NULL
		LIMIT 1`
	var result interfaces.AccessTokenJoinResult
	err := r.qx.GetContext(ctx, &result, q, id, constants.AccessTokenStatusIssued)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &result, err
}
