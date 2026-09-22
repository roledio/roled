package models

import (
	"fmt"
	"html/template"
	"regexp"
)

type GetBrandingRequest struct {
	ProjectID string `uri:"project_id" validate:"required"`
}

type UpdateBrandingRequest struct {
	ProjectID    string  `uri:"project_id" validate:"required"`
	LogoURL      *string `json:"logo_url" validate:"omitempty,http_url,max=2048"`
	FaviconURL   *string `json:"favicon_url" validate:"omitempty,http_url,max=2048"`
	PrimaryColor string  `json:"primary_color" validate:"required,hexcolor,len=7"`
	Rounding     string  `json:"rounding" validate:"required,oneof=sharp small medium large"`
	EnableShadow *bool   `json:"enable_shadow" validate:"required"`
	EnableBorder *bool   `json:"enable_border" validate:"required"`
}

type BrandingDetails struct {
	ProjectID       string  `json:"project_id"`
	SourceProjectID string  `json:"source_project_id"`
	IsDefault       bool    `json:"is_default"`
	LogoURL         *string `json:"logo_url"`
	FaviconURL      *string `json:"favicon_url"`
	PrimaryColor    string  `json:"primary_color"`
	Rounding        string  `json:"rounding"`
	EnableShadow    bool    `json:"enable_shadow"`
	EnableBorder    bool    `json:"enable_border"`
}

var brandingColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// CSS only contains validated colors and fixed tokens, never arbitrary stored CSS.
func (b *BrandingDetails) CSS() template.CSS {
	color := b.PrimaryColor
	if !brandingColorPattern.MatchString(color) {
		color = "#ba8d1c"
	}
	radius := map[string]string{"sharp": "0", "small": "0.25rem", "medium": "0.5rem", "large": "1rem"}[b.Rounding]
	if radius == "" {
		radius = "0.25rem"
	}
	shadow := "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)"
	if !b.EnableShadow {
		shadow = "none"
	}
	border := "0"
	if b.EnableBorder {
		border = "1px solid var(--roled-primary)"
	}
	return template.CSS(fmt.Sprintf(":root{--brand-primary:%s;--brand-radius:%s;--brand-shadow:%s;--brand-border:%s}", color, radius, shadow, border))
}
