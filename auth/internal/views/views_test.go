package views_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/template/html/v3"
	"github.com/roledio/roled/auth/internal/views"
	"github.com/stretchr/testify/assert"
)

func TestTemplates_Load(t *testing.T) {
	engine := html.NewFileSystem(http.FS(views.TemplatesFS), ".html")
	engine.AddFunc("getenv", os.Getenv)

	err := engine.Load()
	assert.NoError(t, err, "templates should load without syntax errors")
}
