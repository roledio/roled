package resource

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

func TestGetResourceDetailsFailuresAndMapping(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "lookup error", "not found", "permissions error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockResourceRepository(t)
			failure := errors.New("offline")
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
				reg.On("ResourceRepository").Return(repo).Once()
				value := &entities.Resource{ID: "id", Name: "Documents", Code: "documents"}
				if stage == "not found" {
					value = nil
				}
				var e error
				if stage == "lookup error" {
					e = failure
				}
				repo.On("FindByIDAndProjectID", ctx, "id", "project").Return(value, e).Once()
				if e == nil && value != nil {
					p := im.NewMockPermissionRepository(t)
					reg.On("PermissionRepository").Return(p).Once()
					var e error
					if stage == "permissions error" {
						e = failure
					}
					p.On("FindByResourceIDsAndSearch", ctx, []string{"id"}, "").Return([]entities.Permission{{ID: "read", ResourceID: "id", Name: "Read", Code: "read"}}, e).Once()
				}
			}
			result, err := NewResourceService(reg, nil).GetResourceDetails(ctx, &models.GetResourceDetailsRequest{ProjectID: "project", ResourceID: "id"})
			if stage == "success" {
				require.NoError(t, err)
				require.Equal(t, "id", result.ID)
				require.Equal(t, "Documents", result.Name)
				require.Len(t, result.Permissions, 1)
				require.Equal(t, "read", result.Permissions[0].ID)
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
