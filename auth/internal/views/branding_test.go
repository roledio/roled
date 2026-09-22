package views_test

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/template/html/v3"
	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/views"
	"github.com/stretchr/testify/require"
)

func TestBrandingOnAllAuthenticationTemplates(t *testing.T) {
	engine := html.NewFileSystem(http.FS(views.TemplatesFS), ".html")
	engine.AddFunc("getenv", os.Getenv)
	require.NoError(t, engine.Load())
	logo, favicon := "https://example.com/brand.png", "https://example.com/favicon.png"
	data := models.TemplateData{
		Branding: &models.BrandingDetails{LogoURL: &logo, FaviconURL: &favicon, PrimaryColor: "#123abc", Rounding: "large", EnableBorder: true},
		Data:     map[string]any{"Project": &entities.Project{Name: "Test project"}, "ProjectSetting": &entities.ProjectSetting{}, "User": &entities.User{}, "Member": &entities.Member{}},
		Flash:    map[string]any{"IsSignup": false}, Map: map[string]any{"Req": models.RenderAuthorizeRequest{}},
	}
	for _, name := range []string{"authorize", "forgot-password", "reset-password", "reset-password-success", "verify-email", "activate-member", "activate-member-success", "activate-project-user", "activate-project-user-success"} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			require.NoError(t, engine.Render(&output, "templates/"+name, &data))
			body := output.String()
			require.Contains(t, body, logo)
			require.Contains(t, body, favicon)
			require.Contains(t, body, "--brand-primary:#123abc")
			require.Contains(t, body, "--brand-radius:1rem")
			require.Contains(t, body, "--brand-shadow:none")
			require.NotContains(t, body, "ZgotmplZ")
		})
	}
}

func TestBrandingCSSRejectsStoredInjection(t *testing.T) {
	b := &models.BrandingDetails{PrimaryColor: "red;}</style><script>alert(1)</script>", Rounding: "injected"}
	css := string(b.CSS())
	require.False(t, strings.Contains(css, "script"))
	require.Contains(t, css, "--brand-primary:#ba8d1c")
	require.Contains(t, css, "--brand-radius:0.25rem")
}
