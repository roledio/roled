package role

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	"github.com/roledio/roled/auth/internal/repositories/interfaces"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateRolePermissionsAndTransaction(t *testing.T) {
	for _, stage := range []string{"success", "empty permissions", "lookup error", "duplicate", "permissions error", "missing permissions", "write error", "links error", "commit error", "find error", "not found", "same code", "no affected", "delete links error"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "account"})
			failure := errors.New("database unavailable")
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			projects, roles := im.NewMockProjectRepository(t), im.NewMockRoleRepository(t)
			permissions, links := im.NewMockPermissionRepository(t), im.NewMockRolePermissionRepository(t)
			reg.On("ProjectRepository").Return(projects).Once()
			projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(&entities.Project{ID: "project"}, nil).Once()
			reg.On("RoleRepository").Return(roles)
			role := &entities.Role{ID: "role", AccountID: "account", ProjectID: "project", Code: "reader"}
			if stage == "same code" {
				role.Code = "editor"
			}
			var findErr error
			if stage == "find error" {
				findErr = failure
			}
			if stage == "not found" {
				role = nil
			}
			roles.On("FindByIDAndProjectID", ctx, "role", "project").Return(role, findErr).Once()
			stop := findErr != nil || role == nil
			if !stop && stage != "same code" {
				var existing *entities.Role
				var lookupErr error
				if stage == "lookup error" {
					lookupErr = failure
				}
				if stage == "duplicate" {
					existing = &entities.Role{ID: "other"}
				}
				roles.On("FindByProjectIDAndCode", ctx, "project", "editor").Return(existing, lookupErr).Once()
				stop = lookupErr != nil || existing != nil
			}
			ids := []string{"read", "write"}
			if stage == "empty permissions" {
				ids = nil
			}
			if !stop && len(ids) > 0 {
				reg.On("PermissionRepository").Return(permissions).Once()
				found := []interfaces.PermissionResource{{ID: "read", Name: "Read", ResourceName: "Documents"}, {ID: "write", Name: "Write", ResourceName: "Documents"}}
				var e error
				if stage == "permissions error" {
					e = failure
				}
				if stage == "missing permissions" {
					found = found[:1]
				}
				permissions.On("FindByIDs", ctx, ids).Return(found, e).Once()
				stop = e != nil || stage == "missing permissions"
			}
			var saved *entities.Role
			if !stop {
				tx.On("RoleRepository").Return(roles).Once()
				var e error
				if stage == "write error" {
					e = failure
				}
				affected := 1
				if stage == "no affected" {
					affected = 0
				}
				roles.On("Update", ctx, mock.Anything).Run(func(a mock.Arguments) {
					saved = a.Get(1).(*entities.Role)
					require.Equal(t, "role", saved.ID)
					require.Equal(t, "editor", saved.Code)
				}).Return(affected, e).Once()
				if e == nil && stage != "no affected" {
					tx.On("RolePermissionRepository").Return(links).Once()
					var deleteErr error
					if stage == "delete links error" {
						deleteErr = failure
					}
					links.On("DeleteByRoleID", ctx, "role").Return(2, deleteErr).Once()
					if deleteErr == nil && len(ids) > 0 {
						var linkErr error
						if stage == "links error" {
							linkErr = failure
						}
						links.On("Create", ctx, []entities.RolePermission{{RoleID: "role", PermissionID: "read"}, {RoleID: "role", PermissionID: "write"}}).Return(linkErr).Once()
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
			result, err := NewRoleService(reg, nil).UpdateRole(ctx, &models.UpdateRoleRequest{ProjectID: "project", RoleID: "role", Code: "EDITOR", Name: "Editor", Description: "Can edit documents", PermissionIDs: ids})
			if stage == "success" || stage == "empty permissions" || stage == "same code" {
				require.NoError(t, err)
				require.Equal(t, saved.ID, result.ID)
				require.Equal(t, "editor", result.Code)
				require.Equal(t, "Editor", result.Name)
				require.Equal(t, "Can edit documents", result.Description)
				require.Len(t, result.Permissions, len(ids))
				if len(ids) > 0 {
					require.Equal(t, "Documents", result.Permissions[0].ResourceName)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}

func TestUpdateRoleProjectGuards(t *testing.T) {
	for _, stage := range []string{"missing account", "foreign project", "project error", "system project"} {
		t.Run(stage, func(t *testing.T) {
			reg := rm.NewMockRegistry(t)
			ctx := context.Background()
			expected := domainerrors.ErrCtxAccountNotFound
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				p := im.NewMockProjectRepository(t)
				reg.On("ProjectRepository").Return(p).Once()
				project := &entities.Project{ID: "project", IsSystem: true}
				var e error
				switch stage {
				case "foreign project":
					project = nil
					expected = domainerrors.ErrProjectNotFound
				case "project error":
					e = errors.New("offline")
					expected = pkgerrors.ErrSystemError
				default:
					expected = pkgerrors.ErrOperationNotAvailable
				}
				p.On("FindByIDAndAccountID", ctx, "project", "account").Return(project, e).Once()
			}
			result, err := NewRoleService(reg, nil).UpdateRole(ctx, &models.UpdateRoleRequest{ProjectID: "project"})
			require.ErrorIs(t, err, expected)
			require.Nil(t, result)
		})
	}
}
