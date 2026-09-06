package mariadb

import (
	"context"
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
	var q = fmt.Sprintf("SELECT * FROM %s WHERE project_id = ?", r.tableName)
	var connections []entities.OAuthConnection
	err := r.qx.SelectContext(ctx, &connections, q, projectID)
	if err != nil {
		return nil, err
	}
	return connections, nil
}
