package role

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetRoleScopePaginationAndFailures(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "count error", "empty", "list error", "related error", "missing settings"} {
		t.Run(stage, func(t *testing.T) {
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockRoleRepository(t)
			ctx := context.Background()
			failure := errors.New("offline")
			req := &models.GetProjectRolesRequest{ProjectID: "project", Search: "read"}
			req.PageNum = 2
			req.PageSize = 5
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				projects := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("RoleRepository").Return(repo).Once()
				n := 7
				if stage == "empty" {
					n = 0
				}
				var e error
				if stage == "count error" {
					e = failure
				}
				repo.On("Count", ctx, req).Return(n, e).Once()
				if e == nil && n > 0 {
					var e error
					if stage == "list error" {
						e = failure
					}
					repo.On("FindAll", ctx, req).Return([]entities.Role{{ID: "one", Name: "Read", Code: "read"}, {ID: "two", Name: "Write", Code: "write"}}, e).Once()
					if e == nil {
						p := im.NewMockProjectSettingRepository(t)
						reg.On("ProjectSettingRepository").Return(p).Once()
						id := "one"
						value := &entities.ProjectSetting{DefaultSignupRoleID: &id}
						if stage == "missing settings" {
							value = nil
						}
						var e error
						if stage == "related error" {
							e = failure
						}
						p.On("FindByProjectID", ctx, "project").Return(value, e).Once()
					}
				}
			}
			result, total, err := NewRoleService(reg, nil).GetRoles(ctx, req)
			switch stage {
			case "success":
				require.NoError(t, err)
				require.Equal(t, 7, total)
				require.Len(t, result, 2)
				require.Equal(t, "Read", result[0].Name)
				require.True(t, *result[0].IsDefaultSignup)
				require.False(t, *result[1].IsDefaultSignup)
			case "empty":
				require.NoError(t, err)
				require.Zero(t, total)
				require.Empty(t, result)
			default:
				require.Error(t, err)
				require.Zero(t, total)
				require.Empty(t, result)
			}
		})
	}
}
