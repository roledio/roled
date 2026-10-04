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

func TestGetRoleDetailsFailuresAndMapping(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "lookup error", "not found", "permissions error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockRoleRepository(t)
			failure := errors.New("offline")
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("RoleRepository").Return(repo).Once()
				value := &entities.Role{ID: "id", Name: "Documents", Code: "documents"}
				if stage == "not found" {
					value = nil
				}
				var e error
				if stage == "lookup error" {
					e = failure
				}
				repo.On("FindByIDAndProjectID", ctx, "id", "project").Return(value, e).Once()
				if e == nil && value != nil {
					p := im.NewMockClientPermissionRepository(t)
					reg.On("ClientPermissionRepository").Return(p).Once()
					var e error
					if stage == "permissions error" {
						e = failure
					}
					p.On("FindByRoleID", ctx, "id").Return([]entities.RolePermission{{PermissionID: "read"}}, e).Once()
				}
			}
			result, err := NewRoleService(reg, nil).GetRoleDetails(ctx, &models.GetRoleDetailsRequest{ProjectID: "project", RoleID: "id"})
			if stage == "success" {
				require.NoError(t, err)
				require.Equal(t, "id", result.ID)
				require.Equal(t, "Documents", result.Name)
				require.Equal(t, []string{"read"}, result.PermissionIDs)
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
