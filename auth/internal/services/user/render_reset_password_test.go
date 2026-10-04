package user

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants/rediskeys"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	cm "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/shomali11/util/xhashes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderResetPasswordValidatesAccountUserProject(t *testing.T) {
	for _, stage := range []string{"success", "redis error", "expired", "user error", "missing user", "inactive user", "account error", "missing account", "inactive account", "project error", "missing project", "inactive project", "flash project", "flash missing project", "flash project error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			cache := cm.NewMockService(t)
			users := im.NewMockUserRepository(t)
			accounts := im.NewMockAccountRepository(t)
			projects := im.NewMockProjectRepository(t)
			failure := errors.New("offline")
			req := &models.RenderResetPasswordRequest{Token: "token"}
			projectID := "project"
			stop := false
			flash := stage == "flash project" || stage == "flash missing project" || stage == "flash project error"
			if flash {
				req.ProjectID = &projectID
			} else {
				var e error
				if stage == "redis error" {
					e = failure
				}
				cache.On("GetData", ctx, rediskeys.ResetPasswordPrefix+":"+xhashes.SHA256("token"), mock.Anything).Run(func(a mock.Arguments) { a.Get(2).(*models.ResetPasswordTokenData).UserID = "user" }).Return(stage != "expired", e).Once()
				stop = e != nil || stage == "expired"
				if !stop {
					reg.On("UserRepository").Return(users).Once()
					u := &entities.User{ID: "user", AccountID: "account", ProjectID: "project", IsActive: stage != "inactive user"}
					var e error
					if stage == "user error" {
						e = failure
					}
					if stage == "missing user" {
						u = nil
					}
					users.On("FindByID", ctx, "user").Return(u, e).Once()
					stop = e != nil || u == nil || stage == "inactive user"
				}
				if !stop {
					reg.On("AccountRepository").Return(accounts).Once()
					a := &entities.Account{ID: "account", IsActive: stage != "inactive account"}
					var e error
					if stage == "account error" {
						e = failure
					}
					if stage == "missing account" {
						a = nil
					}
					accounts.On("FindByID", ctx, "account").Return(a, e).Once()
					stop = e != nil || a == nil || stage == "inactive account"
				}
			}
			if !stop {
				reg.On("ProjectRepository").Return(projects).Once()
				p := &entities.Project{ID: "project", IsActive: stage != "inactive project"}
				var e error
				if stage == "project error" || stage == "flash project error" {
					e = failure
				}
				if stage == "missing project" || stage == "flash missing project" {
					p = nil
				}
				projects.On("FindByID", ctx, "project").Return(p, e).Once()
			}
			result, err := (&userService{registry: reg, redisService: cache}).RenderResetPassword(ctx, req)
			if stage == "success" || stage == "flash project" {
				require.NoError(t, err)
				require.Equal(t, "project", result.Project.ID)
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
