package interfaces

import (
	"context"

	"github.com/roledio/roled/internal/entities"
	"github.com/roledio/roled/internal/models"
)

type MemberRepository interface {
	Create(ctx context.Context, member *entities.Member) error
	FindByAccountIDAndUserID(ctx context.Context, accountID string, userID string) (*entities.Member, error)
	FindAll(ctx context.Context, req *models.GetMembersRequest) ([]MemberUser, error)
	Count(ctx context.Context, req *models.GetMembersRequest) (int, error)
	FindByID(ctx context.Context, id string) (*entities.Member, error)
	Delete(ctx context.Context, member *entities.Member) (int, error)
	Update(ctx context.Context, member *entities.Member) (int, error)
	CountByAccountID(ctx context.Context, accountID string, isAdmin *bool) (int, error)
	DeleteByAccountID(ctx context.Context, accountID string) (int, error)
	FindByIDJoinUser(ctx context.Context, id string) (*MemberUser, error)
}

type MemberUser struct {
	entities.Member
	Email       string  `db:"email"`
	DisplayName string  `db:"display_name"`
	AvatarURL   *string `db:"avatar_url"`
	IsActive    bool    `db:"is_active"`
	IsVerified  bool    `db:"is_verified"`
}
