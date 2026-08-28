package views

import (
	"embed"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/configs"
)

//go:embed templates
var TemplatesFS embed.FS

//go:embed assets
var AssetsFS embed.FS

// RenderTemplate renders a template with automatic minification support based on environment.
// It uses minified versions in non-local environments or when ForceMinified config is enabled.
//
// Parameters:
//   - c: Fiber context
//   - templateName: The template name (e.g., "templates/authorize")
//   - data: Template data to pass to the template
//   - config: Application config to determine environment and minification settings
//
// Returns error from fiber's Render method
func RenderTemplate(c fiber.Ctx, templateName string, data any, config *configs.DefaultConfig) error {
	// Determine if we should use minified version
	useMinified := config.WebUseMinified || !config.IsEnvLocal()

	if useMinified {
		// Convert template name to minified version
		// e.g., "templates/authorize" -> "templates/authorize.min"
		base := filepath.Base(templateName)
		dir := filepath.Dir(templateName)
		minifiedName := filepath.Join(dir, base+".min")
		return c.Render(minifiedName, data)
	}

	return c.Render(templateName, data)
}

// GetAssetPath returns the appropriate asset path based on environment.
// It returns minified asset paths in non-local environments or when ForceMinified config is enabled.
//
// Parameters:
//   - assetPath: The asset path (e.g., "/assets/static/roled.css")
//   - config: Application config to determine environment and minification settings
//
// Returns the appropriate asset path
func GetAssetPath(assetPath string, config *configs.DefaultConfig) string {
	// Determine if we should use minified version
	useMinified := config.WebUseMinified || !config.IsEnvLocal()

	if !useMinified {
		return assetPath
	}

	// Extract the file extension and base name
	ext := filepath.Ext(assetPath)
	base := strings.TrimSuffix(assetPath, ext)

	// Return minified version
	return base + ".min" + ext
}
