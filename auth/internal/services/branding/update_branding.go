package branding

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/services/shared"
	"github.com/roledio/roled/auth/pkg/errors"
)

func (s *service) UpdateBranding(ctx context.Context, req *models.UpdateBrandingRequest) (*models.BrandingDetails, error) {
	_, project, err := shared.ValidateProject(ctx, s.registry, req.ProjectID)
	if err != nil {
		return nil, err
	}
	// Defense in depth for non-HTTP callers, too.
	if !brandingRequestValid(req) {
		return nil, errors.ErrInvalidParams
	}
	logo := req.LogoURL
	// A nil logo follows the project's logo, including later project-logo changes.
	if logo != nil && project.LogoURL != nil && *logo == *project.LogoURL {
		logo = nil
	}
	logo, logoMove, err := s.prepareBrandingAsset(logo, "project-logo")
	if err != nil {
		return nil, err
	}
	favicon, faviconMove, err := s.prepareBrandingAsset(req.FaviconURL, "favicon")
	if err != nil {
		return nil, err
	}
	var moved []brandingAssetMove
	saved := false
	defer func() {
		if saved {
			return
		}
		for i := len(moved) - 1; i >= 0; i-- {
			if err := s.uploadService.Move(ctx, moved[i].to, moved[i].from); err != nil {
				log.WithContext(ctx).Errorw("Failed to restore temporary branding asset", "error", err)
			}
		}
	}()
	for _, move := range []*brandingAssetMove{logoMove, faviconMove} {
		if move == nil {
			continue
		}
		if err := s.uploadService.Move(ctx, move.from, move.to); err != nil {
			return nil, err
		}
		moved = append(moved, *move)
	}
	b := &entities.Branding{
		ProjectID:    project.ID,
		LogoURL:      logo,
		FaviconURL:   favicon,
		PrimaryColor: strings.ToLower(req.PrimaryColor),
		Rounding:     req.Rounding,
		EnableShadow: *req.EnableShadow,
		EnableBorder: *req.EnableBorder,
	}

	_, err = s.registry.BrandingRepository().Upsert(ctx, b)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to upsert branding", "error", err)
		return nil, errors.ErrSystemError.WithError(err)
	}

	saved = true

	// This repository is intentionally uncached; no cached system fallback can go stale.

	if logo == nil {
		logo = project.LogoURL
	}

	return &models.BrandingDetails{
		ProjectID:       project.ID,
		SourceProjectID: project.ID,
		LogoURL:         logo,
		FaviconURL:      favicon,
		PrimaryColor:    b.PrimaryColor,
		Rounding:        b.Rounding,
		EnableShadow:    b.EnableShadow,
		EnableBorder:    b.EnableBorder,
	}, nil
}

func brandingRequestValid(req *models.UpdateBrandingRequest) bool {
	if req.EnableShadow == nil || req.EnableBorder == nil || len(req.PrimaryColor) != 7 || req.PrimaryColor[0] != '#' {
		return false
	}
	for _, c := range req.PrimaryColor[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	switch req.Rounding {
	case "sharp", "small", "medium", "large":
		return true
	}
	return false
}

type brandingAssetMove struct{ from, to string }

func (s *service) prepareBrandingAsset(asset *string, uploadType string) (*string, *brandingAssetMove, error) {
	if asset == nil {
		return nil, nil, nil
	}
	u, err := url.Parse(*asset)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(*asset) > 2048 {
		return nil, nil, errors.ErrInvalidParams
	}
	prefix := strings.TrimRight(s.uploadBaseURL, "/") + "/tmp/"
	if !strings.HasPrefix(*asset, prefix) {
		return asset, nil, nil
	}
	relative := strings.TrimPrefix(*asset, prefix)
	// Only move a single file in the expected upload directory, never arbitrary paths.
	if path.Clean(relative) != relative || strings.ContainsAny(relative, "%\\?#") || path.Dir(relative) != uploadType || path.Base(relative) == "." {
		return nil, nil, errors.ErrInvalidParams
	}
	permanent := strings.TrimRight(s.uploadBaseURL, "/") + "/" + relative
	return &permanent, &brandingAssetMove{from: "tmp/" + relative, to: relative}, nil
}
