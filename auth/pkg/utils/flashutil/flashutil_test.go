package flashutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestSetData_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		SetData(c, map[string]string{"message": "success"})
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	setCookie := resp.Header.Get("Set-Cookie")
	assert.Contains(t, setCookie, "flash_data", "Set-Cookie header should contain flash_data")
}

func TestSetData_NilData(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		SetData(c, nil)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	setCookie := resp.Header.Get("Set-Cookie")
	assert.Empty(t, setCookie, "No Set-Cookie header for nil data")
}

func TestSetData_InvalidData(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		// Data that cannot be marshaled
		SetData(c, make(chan int))
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	setCookie := resp.Header.Get("Set-Cookie")
	assert.Empty(t, setCookie, "No Set-Cookie header for invalid data")
}

func TestReadData_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		var result map[string]string
		err := ReadData(c, &result)
		if err != nil {
			return err
		}
		assert.Equal(t, "test", result["message"])
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Fix for Fiber v3.5.0+ stricter RFC 6265 cookie parsing: double quotes ("), braces ({, })
	// and other non-cookie-octet characters are no longer tolerated in raw Cookie headers
	// by the upgraded fasthttp parser. We must URL-encode the value to match what SetData
	// actually writes to the Set-Cookie header in production.
	req.Header.Set("Cookie", "flash_data="+url.QueryEscape(`{"message":"test"}`))
	resp, err := app.Test(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestReadData_Empty(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		var result map[string]string
		err := ReadData(c, &result)
		assert.NoError(t, err)
		assert.Empty(t, result)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestReadData_InvalidJSON(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		var result map[string]string
		err := ReadData(c, &result)
		assert.Error(t, err)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Fiber v3.5.0 fix: URL-encode the value (space character is disallowed in
	// RFC 6265 cookie-octet). ReadData will QueryUnescape it back and the JSON
	// parse will still correctly fail.
	req.Header.Set("Cookie", "flash_data="+url.QueryEscape("invalid json"))
	resp, err := app.Test(req)
	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestWriteAndRead(t *testing.T) {
	app := fiber.New()

	app.Get("/write", func(c fiber.Ctx) error {
		Write(c, "testkey", "testvalue")
		return c.SendString("written")
	})

	app.Get("/read", func(c fiber.Ctx) error {
		value := Read(c, "testkey")
		return c.SendString(value)
	})

	// Test write
	req1 := httptest.NewRequest("GET", "/write", nil)
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	setCookie1 := resp1.Header.Get("Set-Cookie")
	assert.Contains(t, setCookie1, "flash_testkey", "Set-Cookie should contain flash_testkey")
	assert.Contains(t, setCookie1, "testvalue", "Set-Cookie should contain testvalue")

	// Test read (simulate cookie set)
	req2 := httptest.NewRequest("GET", "/read", nil)
	req2.AddCookie(&http.Cookie{Name: "flash_testkey", Value: "testvalue"})
	resp2, err := app.Test(req2)
	assert.NoError(t, err)

	body, _ := io.ReadAll(resp2.Body)
	assert.Equal(t, "testvalue", string(body))

	// Check clear cookie
	setCookie2 := resp2.Header.Get("Set-Cookie")
	assert.Contains(t, setCookie2, "flash_testkey", "Set-Cookie should contain flash_testkey for clearing")
	assert.Contains(t, setCookie2, "expires=", "Set-Cookie should have expires for clearing")
}

func TestSetDataReadData_RoundTrip(t *testing.T) {
	// This test verifies the actual production flow:
	// Handler 1 calls SetData to write flash cookie → browser receives Set-Cookie
	// → browser sends Cookie on next request → Handler 2 calls ReadData and gets the data back.
	app := fiber.New()

	app.Get("/set", func(c fiber.Ctx) error {
		SetData(c, map[string]string{"message": "hello-world", "key2": "val2"})
		return c.SendString("ok")
	})

	app.Get("/get", func(c fiber.Ctx) error {
		var result map[string]string
		if err := ReadData(c, &result); err != nil {
			return err
		}
		assert.Equal(t, "hello-world", result["message"])
		assert.Equal(t, "val2", result["key2"])
		return c.SendString("ok")
	})

	// Step 1: Trigger SetData and capture the Set-Cookie header
	reqSet := httptest.NewRequest("GET", "/set", nil)
	respSet, err := app.Test(reqSet)
	assert.NoError(t, err)
	assert.Equal(t, 200, respSet.StatusCode)

	setCookie := respSet.Header.Get("Set-Cookie")
	assert.NotEmpty(t, setCookie, "Set-Cookie header should be present")
	assert.Contains(t, setCookie, "flash_data=", "Set-Cookie should contain flash_data")

	// Step 2: Extract cookie name=value (strip attributes like Path, Expires, etc.)
	// Cookie format: "flash_data=<value>; Path=/; ..."
	cookiePart := strings.SplitN(setCookie, ";", 2)[0]

	// Step 3: Request /get with the cookie forwarded
	reqGet := httptest.NewRequest("GET", "/get", nil)
	reqGet.Header.Set("Cookie", cookiePart)
	respGet, err := app.Test(reqGet)
	assert.NoError(t, err)
	assert.Equal(t, 200, respGet.StatusCode)
}
