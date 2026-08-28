package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/repositories/interfaces"
)

type authCodeRepository struct {
	qx interfaces.QueryExecutor
}

func NewAuthCodeRepository(qx interfaces.QueryExecutor) interfaces.AuthCodeRepository {
	return &authCodeRepository{qx: qx}
}

func (r *authCodeRepository) Create(ctx context.Context, authCode *entities.AuthCode) error {
	var q = `INSERT INTO auth_codes (
				id,
				account_id,
				project_id,
				client_id,
				user_id,
				code_hash,
				code_challenge,
				code_challenge_method,
				redirect_uri,
				state,
				expires_at,
				used_at
			) VALUES (
				:id,
				:account_id,
				:project_id,
				:client_id,
				:user_id,
				:code_hash,
				:code_challenge,
				:code_challenge_method,
				:redirect_uri,
				:state,
				:expires_at,
				:used_at
			)`
	_, err := r.qx.NamedExecContext(ctx, q, authCode)
	return err
}

func (r *authCodeRepository) FindByClientIDAndCodeHash(ctx context.Context, clientID string, codeHash string) (*entities.AuthCode, error) {
	var q = "SELECT * FROM auth_codes WHERE client_id = ? AND code_hash = ? LIMIT 1"
	var authCode entities.AuthCode
	err := r.qx.GetContext(ctx, &authCode, q, clientID, codeHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &authCode, err
}

func (r *authCodeRepository) UpdateUsedAuthCode(ctx context.Context, authCode *entities.AuthCode) (int, error) {
	var q = `UPDATE auth_codes SET used_at = NOW(4), updated_at = NOW(4) WHERE id = ?`
	return execOne(ctx, r.qx, q, authCode.ID)
}
