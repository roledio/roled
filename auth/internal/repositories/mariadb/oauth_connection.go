package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type oAuthConnectionRepository struct {
	tableName string
	qx        interfaces.QueryExecutor
}

func NewOAuthConnectionRepository(qx interfaces.QueryExecutor) interfaces.OAuthConnectionRepository {
	return &oAuthConnectionRepository{
		tableName: "oauth_connections",
		qx:        qx,
	}
}

func (r *oAuthConnectionRepository) FindByProjectID(ctx context.Context, projectID string) ([]entities.OAuthConnection, error) {
	var q = fmt.Sprintf("SELECT * FROM %s WHERE project_id = ? ORDER BY provider ASC", r.tableName)
	var connections []entities.OAuthConnection
	err := r.qx.SelectContext(ctx, &connections, q, projectID)
	if err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *oAuthConnectionRepository) FindByProjectIDAndProvider(ctx context.Context, projectID, provider string) (*entities.OAuthConnection, error) {
	var q = fmt.Sprintf("SELECT * FROM %s WHERE project_id = ? AND provider = ?", r.tableName)
	var connection entities.OAuthConnection
	err := r.qx.GetContext(ctx, &connection, q, projectID, provider)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *oAuthConnectionRepository) Create(ctx context.Context, connection *entities.OAuthConnection) error {
	q := fmt.Sprintf(`INSERT INTO %s (
		id, 
		project_id, 
		provider, 
		credential_type, 
		client_id, 
		client_secret_encrypted, 
		scopes, 
		enabled
	) VALUES (
		:id, 
		:project_id, 
		:provider, 
		:credential_type, 
		:client_id, 
		:client_secret_encrypted, 
		:scopes, 
		:enabled
	)`, r.tableName)
	_, err := namedExecOne(ctx, r.qx, q, connection)
	return err
}

func (r *oAuthConnectionRepository) Update(ctx context.Context, connection *entities.OAuthConnection) (int, error) {
	q := fmt.Sprintf(`UPDATE %s SET 
		client_id = :client_id, 
		client_secret_encrypted = :client_secret_encrypted, 
		scopes = :scopes, 
		enabled = :enabled,
		credential_type = :credential_type
		WHERE project_id = :project_id AND provider = :provider`, r.tableName)
	return namedExecOne(ctx, r.qx, q, connection)
}

func (r *oAuthConnectionRepository) Delete(ctx context.Context, projectID, provider string) (int, error) {
	q := fmt.Sprintf("DELETE FROM %s WHERE project_id = ? AND provider = ?", r.tableName)
	return execOne(ctx, r.qx, q, projectID, provider)
}
