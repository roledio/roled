package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/repositories/interfaces"
)

type refreshTokenRepository struct {
	qx interfaces.QueryExecutor
}

func NewRefreshTokenRepository(qx interfaces.QueryExecutor) interfaces.RefreshTokenRepository {
	return &refreshTokenRepository{qx: qx}
}

func (r *refreshTokenRepository) FindByClientIDAndRefreshTokenHash(ctx context.Context, clientID, refreshTokenHash string) (*entities.RefreshToken, error) {
	var q = "SELECT * FROM refresh_tokens WHERE client_id = ? AND refresh_token_hash = ? LIMIT 1"
	var refreshToken entities.RefreshToken
	err := r.qx.GetContext(ctx, &refreshToken, q, clientID, refreshTokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &refreshToken, err
}

func (r *refreshTokenRepository) UpdateUsedRefreshToken(ctx context.Context, refreshToken *entities.RefreshToken) (int, error) {
	var q = "UPDATE refresh_tokens SET status = ?, used_at = NOW(4), updated_at = NOW(4) WHERE id = ?"
	return execOne(ctx, r.qx, q, constants.RefreshTokenStatusUsed, refreshToken.ID)
}

func (r *refreshTokenRepository) UpdateAsRevoked(ctx context.Context, refreshToken *entities.RefreshToken) (int, error) {
	var q = "UPDATE refresh_tokens SET status = ?, revoked_at = NOW(4), updated_at = NOW(4) WHERE id = ?"
	return execOne(ctx, r.qx, q, constants.RefreshTokenStatusRevoked, refreshToken.ID)
}

func (r *refreshTokenRepository) Create(ctx context.Context, refreshToken *entities.RefreshToken) error {
	var q = `INSERT INTO refresh_tokens (
				id, 
				account_id, 
				project_id, 
				client_id, 
				user_id,
				refresh_token_hash, 
				status, 
				expires_in, 
				issued_at
			) VALUES (
				:id,
				:account_id,
				:project_id,
				:client_id,
				:user_id,
				:refresh_token_hash,
				:status,
				:expires_in,
				:issued_at
			)`
	_, err := r.qx.NamedExecContext(ctx, q, refreshToken)
	return err
}
