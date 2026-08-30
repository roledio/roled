package api

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/auth/internal/configs"
	customerrors "github.com/roledio/roled/auth/internal/errors"
	repositorymocks "github.com/roledio/roled/auth/internal/mocks/repositories"
	servicemocks "github.com/roledio/roled/auth/internal/mocks/services"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestPingRoute(t *testing.T) {
	app := fiber.New()
	redisMock := servicemocks.NewMockRedisService(t)
	deps := &Dependencies{
		Redis: redisMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/ping", h.ping)
	req := httptest.NewRequest("GET", "/system/ping", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	assert.Equal(t, "pong", body)
}

func TestGetSystemHealth_AllOK(t *testing.T) {
	app := fiber.New()
	regMock := repositorymocks.NewMockRegistry(t)
	regMock.EXPECT().Ping().Return(nil)
	redisMock := servicemocks.NewMockRedisService(t)
	redisMock.EXPECT().Ping().Return(nil)
	deps := &Dependencies{
		Registry: regMock,
		Redis:    redisMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/health", h.getSystemHealth)
	req := httptest.NewRequest("GET", "/system/health", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	assert.Equal(t, "all systems are operational", body)
}

func TestGetSystemHealth_DBFail(t *testing.T) {
	app := fiber.New()
	// simulate db fail
	regFail := repositorymocks.NewMockRegistry(t)
	regFail.EXPECT().Ping().Return(errors.New("db fail"))
	redisMock := servicemocks.NewMockRedisService(t)
	redisMock.EXPECT().Ping().Return(nil)
	deps := &Dependencies{
		Registry: regFail,
		Redis:    redisMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/health", h.getSystemHealth)
	req := httptest.NewRequest("GET", "/system/health", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	assert.Contains(t, body, "database: db fail")
}

func TestGetSystemHealth_RedisFail(t *testing.T) {
	app := fiber.New()
	regMock := repositorymocks.NewMockRegistry(t)
	regMock.EXPECT().Ping().Return(nil)
	redisFail := servicemocks.NewMockRedisService(t)
	redisFail.EXPECT().Ping().Return(errors.New("redis fail"))
	deps := &Dependencies{
		Registry: regMock,
		Redis:    redisFail,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/health", h.getSystemHealth)
	req := httptest.NewRequest("GET", "/system/health", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	assert.Contains(t, body, "redis: redis fail")
}

func TestGetSystemHealth_BothFail(t *testing.T) {
	app := fiber.New()
	regFail := repositorymocks.NewMockRegistry(t)
	regFail.EXPECT().Ping().Return(errors.New("db fail"))
	redisFail := servicemocks.NewMockRedisService(t)
	redisFail.EXPECT().Ping().Return(errors.New("redis fail"))
	deps := &Dependencies{
		Registry: regFail,
		Redis:    redisFail,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/health", h.getSystemHealth)
	req := httptest.NewRequest("GET", "/system/health", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	assert.Contains(t, body, "database: db fail")
	assert.Contains(t, body, "redis: redis fail")
}

func TestGetSystemInfo_AllOK(t *testing.T) {
	app := fiber.New()
	regMock := repositorymocks.NewMockRegistry(t)
	regMock.EXPECT().Ping().Return(nil)
	redisMock := servicemocks.NewMockRedisService(t)
	redisMock.EXPECT().Ping().Return(nil)
	deps := &Dependencies{
		Registry: regMock,
		Redis:    redisMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/info", h.getSystemInfo)
	req := httptest.NewRequest("GET", "/system/info", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	// parse JSON body and assert fields
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "up", parsed.Get("data.database.status").String())
	assert.Equal(t, "up", parsed.Get("data.redis.status").String())
}

func TestGetSystemInfo_DBFail(t *testing.T) {
	app := fiber.New()
	regFail := repositorymocks.NewMockRegistry(t)
	regFail.EXPECT().Ping().Return(errors.New("db fail"))
	redisMock := servicemocks.NewMockRedisService(t)
	redisMock.EXPECT().Ping().Return(nil)
	deps := &Dependencies{
		Registry: regFail,
		Redis:    redisMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/info", h.getSystemInfo)
	req := httptest.NewRequest("GET", "/system/info", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "down", parsed.Get("data.database.status").String())
	assert.Equal(t, "db fail", parsed.Get("data.database.error").String())
	assert.Equal(t, "up", parsed.Get("data.redis.status").String())
}

func TestGetSystemInfo_RedisFail(t *testing.T) {
	app := fiber.New()
	regMock := repositorymocks.NewMockRegistry(t)
	regMock.EXPECT().Ping().Return(nil)
	redisFail := servicemocks.NewMockRedisService(t)
	redisFail.EXPECT().Ping().Return(errors.New("redis fail"))
	deps := &Dependencies{
		Registry: regMock,
		Redis:    redisFail,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/info", h.getSystemInfo)
	req := httptest.NewRequest("GET", "/system/info", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "up", parsed.Get("data.database.status").String())
	assert.Equal(t, "down", parsed.Get("data.redis.status").String())
	assert.Equal(t, "redis fail", parsed.Get("data.redis.error").String())
}

func TestGetSystemInfo_BothFail(t *testing.T) {
	app := fiber.New()
	regFail := repositorymocks.NewMockRegistry(t)
	regFail.EXPECT().Ping().Return(errors.New("db fail"))
	redisFail := servicemocks.NewMockRedisService(t)
	redisFail.EXPECT().Ping().Return(errors.New("redis fail"))
	deps := &Dependencies{
		Registry: regFail,
		Redis:    redisFail,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/info", h.getSystemInfo)
	req := httptest.NewRequest("GET", "/system/info", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "down", parsed.Get("data.database.status").String())
	assert.Equal(t, "db fail", parsed.Get("data.database.error").String())
	assert.Equal(t, "down", parsed.Get("data.redis.status").String())
	assert.Equal(t, "redis fail", parsed.Get("data.redis.error").String())
}

func TestGetConsoleConfig_Success(t *testing.T) {
	app := fiber.New()
	projectMock := servicemocks.NewMockProjectService(t)
	expectedResponse := &models.GetConsoleConfigResponse{
		ClientID: "console-client-id",
	}
	projectMock.EXPECT().GetConsoleConfig(context.Background()).Return(expectedResponse, nil)
	deps := &Dependencies{
		ProjectService: projectMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/console/config", h.getConsoleConfig)
	req := httptest.NewRequest("GET", "/system/console/config", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.True(t, parsed.Get("success").Bool())
	assert.Equal(t, "console-client-id", parsed.Get("data.client_id").String())
}

func TestGetConsoleConfig_Error(t *testing.T) {
	app := fiber.New()
	projectMock := servicemocks.NewMockProjectService(t)
	projectMock.EXPECT().GetConsoleConfig(context.Background()).Return(nil, customerrors.ErrInvalidProjectCode)
	deps := &Dependencies{
		ProjectService: projectMock,
	}
	h := NewHandler(app, &configs.DefaultConfig{}, deps)
	app.Get("/system/console/config", h.getConsoleConfig)
	req := httptest.NewRequest("GET", "/system/console/config", nil)
	res, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
	bodyBytes, _ := io.ReadAll(res.Body)
	body := string(bodyBytes)
	parsed := gjson.Parse(body)
	assert.False(t, parsed.Get("success").Bool())
	assert.Equal(t, "The specified project code is invalid.", parsed.Get("error.message").String())
	assert.Equal(t, customerrors.ErrInvalidProjectCode.Code, parsed.Get("error.code").String())
}
