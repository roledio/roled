package api

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/constants"
	"github.com/roledio/roled/pkg/utils/responseutil"
)

func (h *handler) ping(c fiber.Ctx) error {
	return c.SendString("pong")
}

func (h *handler) getSystemHealth(c fiber.Ctx) error {
	statuses := []string{}

	// Ping the database server
	err := h.repo.Ping()
	if err != nil {
		statuses = append(statuses, fmt.Sprintf("database: %s", err.Error()))
	}

	// Ping the redis server
	err = h.redisService.Ping()
	if err != nil {
		statuses = append(statuses, fmt.Sprintf("redis: %s", err.Error()))
	}

	status := "all systems are operational"

	if len(statuses) > 0 {
		status = strings.Join(statuses, ", ")
	}

	return c.SendString(status)
}

func (h *handler) getSystemInfo(c fiber.Ctx) error {
	res := models.GetSystemInfoResponse{
		BuildInfo: models.GetCurrentBuildInfo(h.defaultConfig),
		Database:  models.SystemStatus{Status: constants.SystemStatusUp},
		Redis:     models.SystemStatus{Status: constants.SystemStatusUp},
	}

	// Ping the database server
	err := h.repo.Ping()
	if err != nil {
		res.Database = models.SystemStatus{
			Status: constants.SystemStatusDown,
			Error:  err.Error(),
		}
	}

	// Ping the redis server
	err = h.redisService.Ping()
	if err != nil {
		res.Redis = models.SystemStatus{
			Status: constants.SystemStatusDown,
			Error:  err.Error(),
		}
	}

	return responseutil.SendSuccess(c, res)
}

func (h *handler) getConsoleConfig(c fiber.Ctx) error {
	res, err := h.projectService.GetConsoleConfig(c.Context())
	if err != nil {
		return responseutil.SendError(c, err)
	}
	return responseutil.SendSuccess(c, res)
}
