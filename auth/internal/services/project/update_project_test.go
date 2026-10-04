package project

import (
	"context"
	"errors"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	um "github.com/roledio/roled/auth/internal/services/upload/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateProjectTransactionAndFiles(t *testing.T) {
	for _, stage := range []string{"success", "missing account", "system project", "update error", "no affected", "read redirects error", "delete redirects error", "create redirects error", "move error", "delete file error", "commit error", "external logo", "clear logo"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg, tx := rm.NewMockRegistry(t), rm.NewMockRegistry(t)
			projects := im.NewMockProjectRepository(t)
			redirects := im.NewMockRedirectURIRepository(t)
			upload := um.NewMockUploadService(t)
			failure := errors.New("offline")
			base := "https://auth.example/uploads"
			old := base + "/old.png"
			logo := base + "/tmp/new.png"
			active := false
			description := "Updated description"
			project := &entities.Project{ID: "project", Name: "Old", LogoURL: &old, IsSystem: stage == "system project"}
			req := &models.UpdateProjectRequest{ProjectID: "project", Name: "New", Description: &description, LogoURL: &logo, IsActive: &active, RedirectURIs: []models.RedirectURI{{RedirectURI: "https://app.example/callback", LoginURL: "https://app.example/old"}, {RedirectURI: "https://app.example/callback", LoginURL: "https://app.example/login"}}}
			if stage == "external logo" {
				logo = "https://cdn.example/logo.png"
			}
			if stage == "clear logo" {
				req.LogoURL = nil
			}
			if stage != "missing account" {
				ctx = context.WithValue(ctx, constants.CtxAccount, &entities.Account{ID: "account"})
				reg.On("ProjectRepository").Return(projects).Once()
				projects.On("FindByIDAndAccountID", ctx, "project", "account").Return(project, nil).Once()
				if stage != "system project" {
					tx.On("ProjectRepository").Return(projects).Once()
					var e error
					if stage == "update error" {
						e = failure
					}
					n := 1
					if stage == "no affected" {
						n = 0
					}
					projects.On("Update", ctx, mock.Anything).Run(func(a mock.Arguments) {
						v := a.Get(1).(*entities.Project)
						require.Equal(t, "New", v.Name)
						require.False(t, v.IsActive)
						require.Equal(t, &description, v.Description)
					}).Return(n, e).Once()
					stop := e != nil || n == 0
					if !stop {
						tx.On("RedirectURIRepository").Return(redirects).Once()
						var e error
						if stage == "read redirects error" {
							e = failure
						}
						redirects.On("FindByProjectID", ctx, "project").Return([]entities.RedirectURI{{ProjectID: "project", RedirectURI: "https://old.example"}}, e).Once()
						stop = e != nil
					}
					if !stop {
						var e error
						if stage == "delete redirects error" {
							e = failure
						}
						redirects.On("DeleteByProjectID", ctx, "project").Return(1, e).Once()
						stop = e != nil
					}
					if !stop {
						var e error
						if stage == "create redirects error" {
							e = failure
						}
						redirects.On("Create", ctx, mock.Anything).Run(func(a mock.Arguments) {
							rows := a.Get(1).([]entities.RedirectURI)
							require.Len(t, rows, 1)
							require.Equal(t, "project", rows[0].ProjectID)
							require.Equal(t, "https://app.example/login", *rows[0].LoginURL)
						}).Return(e).Once()
						stop = e != nil
					}
					if !stop && stage != "external logo" && stage != "clear logo" {
						var e error
						if stage == "move error" {
							e = failure
						}
						upload.On("Move", ctx, "tmp/new.png", "new.png").Return(e).Once()
						stop = e != nil
					}
					if !stop {
						var e error
						if stage == "delete file error" {
							e = failure
						}
						upload.On("Delete", ctx, "old.png").Return(e).Once()
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
			result, err := (&projectService{registry: reg, uploadService: upload, uploadBaseURL: base}).UpdateProject(ctx, req)
			switch stage {
			case "success", "external logo", "clear logo", "delete file error":
				require.NoError(t, err)
				require.Equal(t, "project", result.ID)
				require.Equal(t, []models.RedirectURI{{RedirectURI: "https://app.example/callback", LoginURL: "https://app.example/login"}}, result.RedirectURIs)
				switch stage {
				case "clear logo":
					require.Nil(t, result.LogoURL)
				case "external logo":
					require.Equal(t, logo, *result.LogoURL)
				default:
					require.Equal(t, base+"/new.png", *result.LogoURL)
				}
			default:
				require.Error(t, err)
				require.Nil(t, result)
			}
		})
	}
}
