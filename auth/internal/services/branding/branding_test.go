package branding

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/constants"
	"github.com/roledio/roled/auth/internal/entities"
	domainerrors "github.com/roledio/roled/auth/internal/errors"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	um "github.com/roledio/roled/auth/internal/services/upload/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBrandingFallbackAndCustom(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(map[bool]string{true: "custom", false: "system fallback"}[custom], func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockBrandingRepository(t)
			projects := im.NewMockProjectRepository(t)
			reg.EXPECT().BrandingRepository().Return(repo)
			logo := "https://example.com/logo.png"
			stored := &entities.Branding{ProjectID: "system", LogoURL: &logo, PrimaryColor: "#2244aa", Rounding: "large", EnableBorder: true}
			if custom {
				stored.ProjectID = "project"
				repo.EXPECT().FindByProjectID(ctx, "project").Return(stored, nil)
			} else {
				reg.EXPECT().ProjectRepository().Return(projects)
				repo.EXPECT().FindByProjectID(ctx, "project").Return(nil, nil)
				projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system"}, nil)
				repo.EXPECT().FindByProjectID(ctx, "system").Return(stored, nil)
			}
			service := NewService(&configs.DefaultConfig{}, reg, nil)
			got, err := service.ResolveBranding(ctx, "project")
			require.NoError(t, err)
			require.Equal(t, !custom, got.IsDefault)
			require.Equal(t, stored.ProjectID, got.SourceProjectID)
			require.Equal(t, "project", got.ProjectID)
			require.Equal(t, stored.PrimaryColor, got.PrimaryColor)
			require.Equal(t, stored.LogoURL, got.LogoURL)
			require.True(t, got.EnableBorder)
			require.False(t, got.EnableShadow)
		})
	}
}

func TestBrandingAccountIsolation(t *testing.T) {
	ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "owner"})
	reg := rm.NewMockRegistry(t)
	projects := im.NewMockProjectRepository(t)
	reg.EXPECT().ProjectRepository().Return(projects)
	projects.EXPECT().FindByIDAndAccountID(ctx, "foreign-project", "owner").Return(nil, nil).Twice()
	service := NewService(&configs.DefaultConfig{}, reg, nil)
	_, err := service.GetBranding(ctx, &models.GetBrandingRequest{ProjectID: "foreign-project"})
	require.ErrorIs(t, err, domainerrors.ErrProjectNotFound)
	_, err = service.UpdateBranding(ctx, &models.UpdateBrandingRequest{ProjectID: "foreign-project"})
	require.ErrorIs(t, err, domainerrors.ErrProjectNotFound)
}

func TestBrandingSaveAndRestoreUploadsOnFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{true: "database failure", false: "save"}[fail], func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constants.CtxAccount, &entities.Account{ID: "owner"})
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockBrandingRepository(t)
			projects := im.NewMockProjectRepository(t)
			uploads := um.NewMockUploadService(t)
			reg.EXPECT().ProjectRepository().Return(projects)
			reg.EXPECT().BrandingRepository().Return(repo)
			projectLogo := "https://example.com/project.png"
			projects.EXPECT().FindByIDAndAccountID(ctx, "project", "owner").Return(&entities.Project{ID: "project", LogoURL: &projectLogo}, nil)
			favicon := "http://localhost/uploads/tmp/favicon/icon.png"
			uploads.EXPECT().Move(ctx, "tmp/favicon/icon.png", "favicon/icon.png").Return(nil)
			var writeErr error
			if fail {
				writeErr = errors.New("database unavailable")
				uploads.EXPECT().Move(ctx, "favicon/icon.png", "tmp/favicon/icon.png").Return(nil)
			}
			repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(b *entities.Branding) bool {
				return b.ProjectID == "project" && b.LogoURL == nil && *b.FaviconURL == "http://localhost/uploads/favicon/icon.png" && !b.EnableShadow && !b.EnableBorder
			})).Return(1, writeErr)
			config := &configs.DefaultConfig{BaseURL: "http://localhost"}
			config.Upload.Driver = constants.UploadDriverLocal
			service := NewService(config, reg, uploads)
			disabled := false
			got, err := service.UpdateBranding(ctx, &models.UpdateBrandingRequest{ProjectID: "project", LogoURL: &projectLogo, FaviconURL: &favicon, PrimaryColor: "#2244AA", Rounding: "sharp", EnableShadow: &disabled, EnableBorder: &disabled})
			if fail {
				require.ErrorIs(t, err, pkgerrors.ErrSystemError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "#2244aa", got.PrimaryColor)
			require.Equal(t, projectLogo, *got.LogoURL)
			require.False(t, got.IsDefault)
		})
	}
}

func TestBrandingAssetRejectsTraversalAndWrongType(t *testing.T) {
	service := &service{uploadBaseURL: "http://localhost/uploads"}
	for _, value := range []string{"javascript:alert(1)", "http://localhost/uploads/tmp/favicon/../secret.png", "http://localhost/uploads/tmp/favicon/%2e%2e/secret.png", "http://localhost/uploads/tmp/project-logo/logo.png"} {
		_, _, err := service.prepareBrandingAsset(&value, "favicon")
		require.ErrorIs(t, err, pkgerrors.ErrInvalidParams, value)
	}
}
