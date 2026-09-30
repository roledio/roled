package authorize

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/entities"
	autherrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGoogleUserCreationFailurePaths(t *testing.T) {
	for _, stage := range []string{"email lookup", "link identity", "signup disabled", "create account", "account lookup", "account missing", "create user", "create identity", "assign role", "create member", "unverified success"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			users := im.NewMockUserRepository(t)
			identities := im.NewMockUserIdentityRepository(t)
			accounts := im.NewMockAccountRepository(t)
			reg.EXPECT().UserRepository().Return(users)
			reg.EXPECT().UserIdentityRepository().Return(identities)
			failure := errors.New("write failed")
			project := &entities.Project{ID: "tenant", AccountID: "owner", IsSystem: stage == "create account"}
			settings := &entities.ProjectSetting{IsSignupEnabled: stage != "signup disabled"}
			info := &models.GoogleUserInfo{Sub: "google-subject", Email: "member@example.com", Name: "Member", Picture: "https://example.com/avatar.png"}
			var existing *entities.User
			var lookupErr error
			if stage == "email lookup" {
				lookupErr = failure
			}
			if stage == "link identity" {
				existing = &entities.User{ID: "existing", ProjectID: project.ID}
			}
			users.EXPECT().FindByProjectIDAndEmail(ctx, "tenant", info.Email).Return(existing, lookupErr)
			if stage == "link identity" {
				identities.EXPECT().Create(ctx, mock.MatchedBy(func(i *entities.UserIdentity) bool {
					return i.ID != "" && i.UserID == "existing" && i.ProjectID == "tenant" && i.Provider == "google" && i.ProviderUserID == info.Sub
				})).Return(failure)
			}
			var created *entities.User
			if stage != "email lookup" && stage != "link identity" && stage != "signup disabled" {
				reg.EXPECT().AccountRepository().Return(accounts)
				if stage == "create account" {
					accounts.EXPECT().Create(ctx, mock.MatchedBy(func(a *entities.Account) bool { return a.ID != "" && a.Name == info.Name && a.IsActive && !a.IsSystem })).Return(failure)
				} else {
					account := &entities.Account{ID: "owner"}
					var accountErr error
					if stage == "account lookup" {
						accountErr = failure
					}
					if stage == "account missing" {
						account = nil
					}
					accounts.EXPECT().FindByID(ctx, "owner").Return(account, accountErr)
					if account != nil && accountErr == nil {
						var userErr error
						if stage == "create user" {
							userErr = failure
						}
						users.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, u *entities.User) {
							created = u
							require.NotEmpty(t, u.ID)
							require.Equal(t, "tenant", u.ProjectID)
							require.Equal(t, "owner", u.AccountID)
							require.Equal(t, info.Email, *u.Email)
							require.Equal(t, info.Name, u.DisplayName)
							require.Equal(t, info.Picture, *u.AvatarURL)
							require.True(t, u.IsActive)
							require.Nil(t, u.EmailVerifiedAt)
						}).Return(userErr)
						if userErr == nil {
							var identityErr error
							if stage == "create identity" {
								identityErr = failure
							}
							identities.EXPECT().Create(ctx, mock.MatchedBy(func(i *entities.UserIdentity) bool {
								return created != nil && i.UserID == created.ID && i.ProjectID == "tenant" && i.Provider == "google" && i.ProviderUserID == info.Sub
							})).Return(identityErr)
							if stage == "assign role" {
								role := "default-role"
								settings.DefaultSignupRoleID = &role
								roles := im.NewMockUserRoleRepository(t)
								reg.EXPECT().UserRoleRepository().Return(roles)
								roles.EXPECT().Create(ctx, mock.MatchedBy(func(r *entities.UserRole) bool { return r.UserID == created.ID && r.RoleID == role })).Return(failure)
							}
							if stage == "create member" || stage == "unverified success" {
								members := im.NewMockMemberRepository(t)
								reg.EXPECT().MemberRepository().Return(members)
								var memberErr error
								if stage == "create member" {
									memberErr = failure
								}
								members.EXPECT().Create(ctx, mock.MatchedBy(func(m *entities.Member) bool {
									return m.ID != "" && m.AccountID == "owner" && m.UserID == created.ID && m.IsAdmin
								})).Return(memberErr)
							}
						}
					}
				}
			}
			got, err := (&authorizeService{}).createOrUpdateUserFromGoogle(ctx, reg, project, settings, info)
			if stage == "unverified success" {
				require.NoError(t, err)
				require.Equal(t, created, got)
				return
			}
			require.Nil(t, got)
			if stage == "signup disabled" {
				require.ErrorIs(t, err, autherrors.ErrUnableToProcessSignup)
			} else {
				require.ErrorIs(t, err, pkgerrors.ErrSystemError)
			}
		})
	}
}
