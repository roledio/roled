package branding

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	im "github.com/roledio/roled/auth/internal/repositories/interfaces/mocks"
	rm "github.com/roledio/roled/auth/internal/repositories/mocks"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestResolveBrandingFailures(t *testing.T) {
	for _, stage := range []string{"project branding", "system lookup", "missing system", "system branding"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockBrandingRepository(t)
			projects := im.NewMockProjectRepository(t)
			reg.EXPECT().BrandingRepository().Return(repo)
			failure := errors.New("database unavailable")
			switch stage {
			case "project branding":
				repo.EXPECT().FindByProjectID(ctx, "tenant").Return(nil, failure)
			default:
				// No custom branding for tenant → fall through to system lookup.
				repo.EXPECT().FindByProjectID(ctx, "tenant").Return(nil, nil)
				reg.EXPECT().ProjectRepository().Return(projects)
				switch stage {
				case "system lookup":
					projects.EXPECT().FindSystem(ctx).Return(nil, failure)
				case "missing system":
					projects.EXPECT().FindSystem(ctx).Return(nil, nil)
				case "system branding":
					projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system"}, nil)
					repo.EXPECT().FindByProjectID(ctx, "system").Return(nil, failure)
				}
			}
			got, err := NewService(&configs.DefaultConfig{}, reg, nil).ResolveBranding(ctx, &entities.Project{ID: "tenant"})
			require.Nil(t, got)
			require.ErrorIs(t, err, pkgerrors.ErrSystemError)
		})
	}
}

func TestResolveBrandingWithoutStoredSystemSettings(t *testing.T) {
	ctx := context.Background()
	reg := rm.NewMockRegistry(t)
	repo := im.NewMockBrandingRepository(t)
	projects := im.NewMockProjectRepository(t)
	tenantLogo := "https://example.com/tenant.png"
	reg.EXPECT().BrandingRepository().Return(repo)
	reg.EXPECT().ProjectRepository().Return(projects)
	repo.EXPECT().FindByProjectID(ctx, "tenant").Return(nil, nil)
	projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system"}, nil)
	repo.EXPECT().FindByProjectID(ctx, "system").Return(nil, nil)
	// Logo comes from the project entity passed to ResolveBranding (not the system project).
	got, err := NewService(&configs.DefaultConfig{}, reg, nil).ResolveBranding(ctx, &entities.Project{ID: "tenant", LogoURL: &tenantLogo})
	require.NoError(t, err)
	require.Equal(t, &models.BrandingDetails{ProjectID: "tenant", SourceProjectID: "system", IsDefault: true, LogoURL: &tenantLogo, PrimaryColor: "#ba8d1c", Rounding: "small", EnableShadow: true}, got)
}

func TestResolveBrandingWithoutStoredSystemSettingsNoLogo(t *testing.T) {
	ctx := context.Background()
	reg := rm.NewMockRegistry(t)
	repo := im.NewMockBrandingRepository(t)
	projects := im.NewMockProjectRepository(t)
	reg.EXPECT().BrandingRepository().Return(repo)
	reg.EXPECT().ProjectRepository().Return(projects)
	repo.EXPECT().FindByProjectID(ctx, "tenant").Return(nil, nil)
	projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system"}, nil)
	repo.EXPECT().FindByProjectID(ctx, "system").Return(nil, nil)
	// Project has no logo → LogoURL stays nil.
	got, err := NewService(&configs.DefaultConfig{}, reg, nil).ResolveBranding(ctx, &entities.Project{ID: "tenant"})
	require.NoError(t, err)
	require.Equal(t, &models.BrandingDetails{ProjectID: "tenant", SourceProjectID: "system", IsDefault: true, PrimaryColor: "#ba8d1c", Rounding: "small", EnableShadow: true}, got)
}

func TestResolveBrandingUsesStoredSettings(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "custom branding", true: "system fallback"}[fallback], func(t *testing.T) {
			ctx := context.Background()
			reg := rm.NewMockRegistry(t)
			repo := im.NewMockBrandingRepository(t)
			projects := im.NewMockProjectRepository(t)
			reg.EXPECT().BrandingRepository().Return(repo)
			tenantLogo := "https://example.com/tenant.png"
			sourceID := "tenant"
			favicon := "https://example.com/icon.png"
			stored := &entities.Branding{ProjectID: "tenant", FaviconURL: &favicon, PrimaryColor: "#112233", Rounding: "sharp", EnableBorder: true}
			if fallback {
				sourceID = "system"
				stored.ProjectID = sourceID
				// No custom branding for tenant → fall through to system.
				repo.EXPECT().FindByProjectID(ctx, "tenant").Return(nil, nil)
				reg.EXPECT().ProjectRepository().Return(projects)
				projects.EXPECT().FindSystem(ctx).Return(&entities.Project{ID: "system"}, nil)
			}
			repo.EXPECT().FindByProjectID(ctx, sourceID).Return(stored, nil)
			// Logo always comes from the project entity passed to ResolveBranding.
			got, err := NewService(&configs.DefaultConfig{}, reg, nil).ResolveBranding(ctx, &entities.Project{ID: "tenant", LogoURL: &tenantLogo})
			require.NoError(t, err)
			require.Equal(t, &models.BrandingDetails{
				ProjectID:       "tenant",
				SourceProjectID: sourceID,
				IsDefault:       fallback,
				LogoURL:         &tenantLogo,
				FaviconURL:      &favicon,
				PrimaryColor:    "#112233",
				Rounding:        "sharp",
				EnableBorder:    true,
			}, got)
		})
	}
}

func TestResolveBrandingStoredLogoFallsBackToBrandingRecord(t *testing.T) {
	// When the project has no logo but the stored branding record has one, use the branding logo.
	ctx := context.Background()
	reg := rm.NewMockRegistry(t)
	repo := im.NewMockBrandingRepository(t)
	reg.EXPECT().BrandingRepository().Return(repo)
	brandingLogo := "https://example.com/branding-logo.png"
	favicon := "https://example.com/icon.png"
	stored := &entities.Branding{ProjectID: "tenant", LogoURL: &brandingLogo, FaviconURL: &favicon, PrimaryColor: "#112233", Rounding: "sharp"}
	repo.EXPECT().FindByProjectID(ctx, "tenant").Return(stored, nil)
	// Project has no logo → result.LogoURL comes from stored branding record.
	got, err := NewService(&configs.DefaultConfig{}, reg, nil).ResolveBranding(ctx, &entities.Project{ID: "tenant"})
	require.NoError(t, err)
	require.Equal(t, &models.BrandingDetails{
		ProjectID:       "tenant",
		SourceProjectID: "tenant",
		LogoURL:         &brandingLogo,
		FaviconURL:      &favicon,
		PrimaryColor:    "#112233",
		Rounding:        "sharp",
	}, got)
}
