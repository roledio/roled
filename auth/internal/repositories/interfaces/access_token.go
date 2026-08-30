package interfaces

import (
	"context"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
)

type AccessTokenRepository interface {
	Create(ctx context.Context, token *entities.AccessToken) error
	FindByID(ctx context.Context, id string) (*entities.AccessToken, error)
	UpdateAsIssued(ctx context.Context, token *entities.AccessToken) (int, error)
	UpdateAsRevoked(ctx context.Context, id string) (int, error)
	DeleteByAccountID(ctx context.Context, accountID string) (int, error)
	DeleteByUserID(ctx context.Context, userID string) (int, error)
	DeleteByProjectID(ctx context.Context, projectID string) (int, error)
	DeleteByClientID(ctx context.Context, clientID string) (int, error)
	FindByIDJoin(ctx context.Context, id string) (*AccessTokenJoinResult, error)
}

type AccessTokenJoinResult struct {
	ID        string    `db:"id"`
	IssuedAt  time.Time `db:"issued_at"`
	ExpiresIn int       `db:"expires_in"`

	ProjectID          string  `db:"project_id"`
	ProjectName        string  `db:"project_name"`
	ProjectDescription string  `db:"project_description"`
	ProjectLogoURL     *string `db:"project_logo_url"`

	ClientID          string `db:"client_id"`
	ClientName        string `db:"client_name"`
	ClientDescription string `db:"client_description"`

	UserID             *string `db:"user_id"`
	UserDisplayName    *string `db:"user_display_name"`
	UserEmail          *string `db:"user_email"`
	UserExternalUserID *string `db:"user_external_user_id"`
	UserAvatarURL      *string `db:"user_avatar_url"`

	RoleID          *string `db:"role_id"`
	RoleCode        *string `db:"role_code"`
	RoleName        *string `db:"role_name"`
	RoleDescription *string `db:"role_description"`
}
