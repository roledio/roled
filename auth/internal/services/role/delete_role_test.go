package role

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeleteRoleTransaction(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "system project", "lookup error", "missing role", "delete permissions error", "delete role error", "no affected", "commit error"} {
		t.Run(stage, func(t *testing.T) {
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			roles := im.NewMockRoleRepository(t)
			links := im.NewMockRolePermissionRepository(t)
			ctx := context.Background()
			failure := errors.New("offline")
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project", IsSystem: stage == "system project"}, nil).Once()
				if stage != "system project" {
					reg.On("RoleRepository").Return(roles).Once()
					role := &entities.Role{ID: "role", ProjectID: "project"}
					var e error
					if stage == "lookup error" {
						e = failure
					}
					if stage == "missing role" {
						role = nil
					}
					roles.On("FindByIDAndProjectID", ctx, "role", "project").Return(role, e).Once()
					if e == nil && role != nil {
						tx.On("RolePermissionRepository").Return(links).Once()
						var de error
						if stage == "delete permissions error" {
							de = failure
						}
						links.On("DeleteByRoleID", ctx, "role").Return(2, de).Once()
						if de == nil {
							tx.On("RoleRepository").Return(roles).Once()
							var re error
							if stage == "delete role error" {
								re = failure
							}
							n := 1
							if stage == "no affected" {
								n = 0
							}
							roles.On("DeleteByID", ctx, "role").Return(n, re).Once()
						}
						reg.On("Tx", mock.Anything).Return(func(fn func(repositories.Registry) error) error {
							if e := fn(tx); e != nil {
								return e
							}
							if stage == "commit error" {
								return failure
							}
							return nil
						}).Once()
					}
				}
			}
			err := NewRoleService(reg, nil).DeleteRole(ctx, &models.DeleteRoleRequest{ProjectID: "project", RoleID: "role"})
			if stage == "success" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
