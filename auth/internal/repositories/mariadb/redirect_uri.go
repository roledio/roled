package mariadb

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
)

type redirectURIRepository struct {
	qx interfaces.QueryExecutor
}

func NewRedirectURIRepository(qx interfaces.QueryExecutor) interfaces.RedirectURIRepository {
	return &redirectURIRepository{qx: qx}
}

func (r *redirectURIRepository) FindByProjectIDAndRedirectURI(ctx context.Context, projectID string, redirectURI string) (*entities.RedirectURI, error) {
	var q = "SELECT * FROM redirect_uris WHERE project_id = ? AND redirect_uri = ?"
	var result entities.RedirectURI
	err := r.qx.GetContext(ctx, &result, q, projectID, redirectURI)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &result, err
}

func (r *redirectURIRepository) FindByProjectID(ctx context.Context, projectID string) ([]entities.RedirectURI, error) {
	var q = "SELECT * FROM redirect_uris WHERE project_id = ?"
	var results []entities.RedirectURI
	err := r.qx.SelectContext(ctx, &results, q, projectID)
	return results, err
}

func (r *redirectURIRepository) Create(ctx context.Context, redirectURIs []entities.RedirectURI) error {
	if len(redirectURIs) == 0 {
		return nil
	}

	builder := sq.
		Insert("redirect_uris").
		Columns("project_id", "redirect_uri", "login_url")

	for _, redirectURI := range redirectURIs {
		builder = builder.Values(
			redirectURI.ProjectID,
			redirectURI.RedirectURI,
			redirectURI.LoginURL,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = exec(ctx, r.qx, query, args...)
	return err
}

func (r *redirectURIRepository) DeleteByProjectID(ctx context.Context, projectID string) (int, error) {
	q := `DELETE FROM redirect_uris WHERE project_id = ?`
	return exec(ctx, r.qx, q, projectID)
}
