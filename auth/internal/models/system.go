package models

import "github.com/roledio/roled/auth/pkg/models"

type GetSystemInfoResponse struct {
	models.BuildInfo
	Database SystemStatus `json:"database"`
	Redis    SystemStatus `json:"redis"`
}

type SystemStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type GetConsoleConfigResponse struct {
	ClientID string `json:"client_id"`
}
