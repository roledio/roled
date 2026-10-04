package resource

import (
	"context"
	"encoding/json"
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

func TestCreateResourceAtomicPermissions(t *testing.T) {
	for _, stage := range []string{"success", "empty permissions", "missing account", "system project", "lookup error", "duplicate", "write error", "permissions error", "commit error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			resources := im.NewMockResourceRepository(t)
			permissions := im.NewMockPermissionRepository(t)
			failure := errors.New("offline")
			var req models.CreateResourceRequest
			require.NoError(t, json.Unmarshal([]byte(`{"name":"Documents","code":"documents","description":"Project files","permissions":[{"name":"Read","code":"read","description":"Read files"}]}`), &req))
			req.ProjectID = "project"
			if stage == "empty permissions" {
				req.Permissions = nil
			}
			var saved *entities.Resource
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project", IsSystem: stage == "system project"}, nil).Once()
				if stage != "system project" {
					reg.On("ResourceRepository").Return(resources)
					var existing *entities.Resource
					var lookupErr error
					if stage == "lookup error" {
						lookupErr = failure
					}
					if stage == "duplicate" {
						existing = &entities.Resource{ID: "other"}
					}
					resources.On("FindByProjectIDAndCode", ctx, "project", "documents").Return(existing, lookupErr).Once()
					stop := lookupErr != nil || existing != nil
					if !stop {
						tx.On("ResourceRepository").Return(resources).Once()
						var e error
						if stage == "write error" {
							e = failure
						}
						n := 1
						if stage == "no affected" {
							n = 0
						}
						resources.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
							rows := a.Get(1).([]entities.Resource)
							require.Len(t, rows, 1)
							saved = &rows[0]
							require.NotEmpty(t, saved.ID)
							require.Equal(t, "account", saved.AccountID)
							require.Equal(t, "project", saved.ProjectID)
						}).Return(n, e).Once()
						if e == nil && n > 0 {
							if len(req.Permissions) > 0 {
								tx.On("PermissionRepository").Return(permissions).Once()
								var pe error
								if stage == "permissions error" {
									pe = failure
								}
								permissions.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
									rows := a.Get(1).([]entities.Permission)
									require.Len(t, rows, 1)
									require.Equal(t, saved.ID, rows[0].ResourceID)
									require.NotEmpty(t, rows[0].ID)
									require.Equal(t, "read", rows[0].Code)
									require.False(t, rows[0].IsDefault)
								}).Return(1, pe).Once()
							}
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
			result, err := NewResourceService(reg, nil).CreateResource(ctx, &req)
			if stage == "success" || stage == "empty permissions" || stage == "same code" {
				require.NoError(t, err)
				require.Equal(t, saved.ID, result.ID)
				require.Equal(t, "Documents", result.Name)
				require.Equal(t, "documents", result.Code)
				require.Equal(t, "Project files", *result.Description)
				require.False(t, result.IsDefault)
				require.Len(t, result.Permissions, len(req.Permissions))
				if len(req.Permissions) > 0 {
					require.NotEmpty(t, result.Permissions[0].ID)
					require.Equal(t, "read", result.Permissions[0].Code)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
