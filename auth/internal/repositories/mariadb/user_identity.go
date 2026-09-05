package mariadb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type userIdentityRepository struct {
	qx interfaces.QueryExecutor
}

func NewUserIdentityRepository(qx interfaces.QueryExecutor) interfaces.UserIdentityRepository {
	return &userIdentityRepository{qx: qx}
}

func (r *userIdentityRepository) Create(ctx context.Context, userIdentity *entities.UserIdentity) error {
	var q = `INSERT INTO user_identities (
				id,
				user_id,
				provider,
				provider_user_id
			) VALUES (
				:id,
				:user_id,
				:provider,
				:provider_user_id
			)`
	_, err := r.qx.NamedExecContext(ctx, q, userIdentity)
	return err
}

func (r *userIdentityRepository) FindByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*entities.UserIdentity, error) {
	var q = "SELECT * FROM user_identities WHERE provider = ? AND provider_user_id = ? AND deleted_at IS NULL LIMIT 1"
	var userIdentity entities.UserIdentity
	err := r.qx.GetContext(ctx, &userIdentity, q, provider, providerUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &userIdentity, err
}

func (r *userIdentityRepository) FindByUserID(ctx context.Context, userID string) ([]*entities.UserIdentity, error) {
	var q = "SELECT * FROM user_identities WHERE user_id = ? AND deleted_at IS NULL"
	var userIdentities []*entities.UserIdentity
	err := r.qx.SelectContext(ctx, &userIdentities, q, userID)
	if err != nil {
		return nil, err
	}
	return userIdentities, nil
}

func (r *userIdentityRepository) DeleteByID(ctx context.Context, id string) (int, error) {
	var q = `UPDATE user_identities SET
				deleted_at = NOW(4),
				updated_at = NOW(4)
			WHERE id = ? AND deleted_at IS NULL`
	return execOne(ctx, r.qx, q, id)
}
