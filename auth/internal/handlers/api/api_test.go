package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/configs"
	repositorymocks "github.com/roledio/roled/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/internal/mocks/services"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	app := fiber.New()
	regMock := repositorymocks.NewMockRegistry(t)
	redisMock := servicemocks.NewMockRedisService(t)
	deps := &Dependencies{
		Registry: regMock,
		Redis:    redisMock,
		// Other fields can be nil for this test
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	h.SetupRoutes()
	// Test that routes are set up by testing one of them
	req := httptest.NewRequest("GET", "/system/ping", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}
