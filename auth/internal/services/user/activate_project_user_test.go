package user

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	redismocks "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/roledio/roled/auth/pkg/utils/passwordutil"
	"github.com/shomali11/util/xhashes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRenderActivateProjectUserTokenAndLookup(t *testing.T) {
	for _, stage := range []string{"success", "redis error", "expired", "user error", "missing user", "project error", "missing project"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			cache := redismocks.NewMockService(t)
			users := im.NewMockUserRepository(t)
			projects := im.NewMockProjectRepository(t)
			failure := errors.New("offline")
			login := "https://app.example/login"
			key := rediskeys.ActivateProjectUserPrefix + ":" + xhashes.SHA256("token")
			var e error
			if stage == "redis error" {
				e = failure
			}
			cache.On("GetData", ctx, key, mock.Anything).Run(func(a mock.Arguments) {
				*a.Get(2).(*models.ActivateProjectUserTokenData) = models.ActivateProjectUserTokenData{UserID: "user", LoginURL: &login}
			}).Return(stage != "expired", e).Once()
			if e == nil && stage != "expired" {
				reg.On("UserRepository").Return(users).Once()
				u := &entities.User{ID: "user", ProjectID: "project"}
				var e error
				if stage == "user error" {
					e = failure
				}
				if stage == "missing user" {
					u = nil
				}
				users.On("FindByID", ctx, "user").Return(u, e).Once()
				if e == nil && u != nil {
					reg.On("ProjectRepository").Return(projects).Once()
					p := &entities.Project{ID: "project"}
					var e error
					if stage == "project error" {
						e = failure
					}
					if stage == "missing project" {
						p = nil
					}
					projects.On("FindByID", ctx, "project").Return(p, e).Once()
				}
			}
			res, err := (&userService{registry: reg, redisService: cache}).RenderActivateProjectUser(ctx, &models.RenderActivateProjectUserRequest{Token: "token"})
			if stage == "success" {
				require.NoError(t, err)
				require.Equal(t, "user", res.User.ID)
				require.Equal(t, "project", res.Project.ID)
				require.Equal(t, &login, res.LoginURL)
			} else {
				require.Error(t, err)
				require.Nil(t, res)
			}
		})
	}
}
func TestSubmitActivateProjectUserAtomicState(t *testing.T) {
	for _, stage := range []string{"success", "redis error", "expired", "user error", "missing user", "already active", "update error", "commit error", "token delete error", "project error", "missing project"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			cache := redismocks.NewMockService(t)
			users := im.NewMockUserRepository(t)
			projects := im.NewMockProjectRepository(t)
			failure := errors.New("offline")
			login := "https://app.example/login"
			key := rediskeys.ActivateProjectUserPrefix + ":" + xhashes.SHA256("token")
			password := "secret passphrase"
			before := time.Now()
			var e error
			if stage == "redis error" {
				e = failure
			}
			cache.On("GetData", ctx, key, mock.Anything).Run(func(a mock.Arguments) {
				*a.Get(2).(*models.ActivateProjectUserTokenData) = models.ActivateProjectUserTokenData{UserID: "user", LoginURL: &login}
			}).Return(stage != "expired", e).Once()
			if e == nil && stage != "expired" {
				reg.On("UserRepository").Return(users).Once()
				hash := "hash"
				u := &entities.User{ID: "user", ProjectID: "project", IsActive: stage == "already active"}
				if stage == "already active" {
					u.PasswordHash = &hash
				}
				var e error
				if stage == "user error" {
					e = failure
				}
				if stage == "missing user" {
					u = nil
				}
				users.On("FindByID", ctx, "user").Return(u, e).Once()
				if e == nil && u != nil && stage != "already active" {
					tx.On("UserRepository").Return(users).Once()
					var ue error
					if stage == "update error" {
						ue = failure
					}
					users.On("Update", ctx, mock.Anything).Run(func(a mock.Arguments) {
						saved := a.Get(1).(*entities.User)
						require.True(t, saved.IsActive)
						require.Equal(t, "Invited User", saved.DisplayName)
						require.True(t, passwordutil.IsValidPassword(password, *saved.PasswordHash))
						require.NotNil(t, saved.EmailVerifiedAt)
						require.False(t, saved.EmailVerifiedAt.Before(before))
						require.Equal(t, "user", saved.ID)
					}).Return(1, ue).Once()
					reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
						if e := fn(tx); e != nil {
							return e
						}
						if stage == "commit error" {
							return failure
						}
						return nil
					}).Once()
					if ue == nil && stage != "commit error" {
						var de error
						if stage == "token delete error" {
							de = failure
						}
						cache.On("DeleteWithContext", ctx, key).Return(de).Once()
						reg.On("ProjectRepository").Return(projects).Once()
						p := &entities.Project{ID: "project"}
						var e error
						if stage == "project error" {
							e = failure
						}
						if stage == "missing project" {
							p = nil
						}
						projects.On("FindByID", ctx, "project").Return(p, e).Once()
					}
				}
			}
			res, err := (&userService{registry: reg, redisService: cache}).SubmitActivateProjectUser(ctx, &models.SubmitActivateProjectUserRequest{Token: "token", Password: password, DisplayName: "Invited User"})
			if stage == "success" || stage == "token delete error" {
				require.NoError(t, err)
				require.Equal(t, "user", res.UserID)
				require.Equal(t, "project", res.Project.ID)
				require.Equal(t, &login, res.LoginURL)
			} else {
				require.Error(t, err)
				require.Nil(t, res)
			}
		})
	}
}
