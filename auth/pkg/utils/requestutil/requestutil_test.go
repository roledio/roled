package requestutil

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestBindAndValidate(t *testing.T) {
	app := fiber.New()

	// Test case 1: Valid request
	app.Post("/test", func(c fiber.Ctx) error {
		type TestReq struct {
			Name string `json:"name" validate:"required"`
			Age  int    `json:"age" validate:"required,min=1"`
		}
		var req TestReq
		err := BindAndValidate(c, &req)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"name": req.Name, "age": req.Age})
	})

	// Valid request
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"name":"John","age":25}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test case 2: Invalid request (missing required field)
	req2 := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"age":25}`)))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp2.StatusCode) // Should return 400 for validation error

	// Test case 3: Invalid JSON
	req3 := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`invalid json`)))
	req3.Header.Set("Content-Type", "application/json")
	resp3, err := app.Test(req3)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp3.StatusCode)

	// Test case 4: Valid with query params
	app.Get("/test-query", func(c fiber.Ctx) error {
		type TestReq struct {
			Name string `query:"name" validate:"required"`
			Age  int    `query:"age" validate:"required,min=1"`
		}
		var req TestReq
		err := BindAndValidate(c, &req)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"name": req.Name, "age": req.Age})
	})

	req4 := httptest.NewRequest("GET", "/test-query?name=Jane&age=30", nil)
	resp4, err := app.Test(req4)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp4.StatusCode)
}
