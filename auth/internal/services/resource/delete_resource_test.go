package resource

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

func TestDeleteResourceGuardsAndTransaction(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "system project", "lookup error", "missing resource", "default resource", "permissions error", "delete error", "no affected", "commit error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			resources := im.NewMockResourceRepository(t)
			permissions := im.NewMockPermissionRepository(t)
			failure := errors.New("offline")
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				projects := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project", IsSystem: stage == "system project"}, nil).Once()
				if stage != "system project" {
					reg.On("ResourceRepository").Return(resources).Once()
					r := &entities.Resource{ID: "resource", ProjectID: "project", IsDefault: stage == "default resource"}
					var e error
					if stage == "lookup error" {
						e = failure
					}
					if stage == "missing resource" {
						r = nil
					}
					resources.On("FindByIDAndProjectID", ctx, "resource", "project").Return(r, e).Once()
					if e == nil && r != nil && !r.IsDefault {
						tx.On("PermissionRepository").Return(permissions).Once()
						var e error
						if stage == "permissions error" {
							e = failure
						}
						permissions.On("DeleteByResourceID", ctx, "resource").Return(2, e).Once()
						if e == nil {
							tx.On("ResourceRepository").Return(resources).Once()
							var e error
							if stage == "delete error" {
								e = failure
							}
							n := 1
							if stage == "no affected" {
								n = 0
							}
							resources.On("Delete", ctx, r).Return(n, e).Once()
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
			err := NewResourceService(reg, nil).DeleteResource(ctx, &models.DeleteResourceRequest{ProjectID: "project", ResourceID: "resource"})
			if stage == "success" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
