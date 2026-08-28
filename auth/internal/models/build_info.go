package models

import (
	"time"

	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/pkg/constants"
	pkgmodels "github.com/roledio/roled/pkg/models"
)

var (
	ProjectName    = "roled"
	AppName        = "Roled"
	AppVersion     = "" // The real value will be set from -ldflags (during build time)
	CommitHash     = "" // The real value will be set from -ldflags (during build time)
	BuildTimestamp = "" // The real value will be set from -ldflags (during build time)
	StartTimestamp = time.Now()
)

func GetCurrentBuildInfo(defaultConfig *configs.DefaultConfig) pkgmodels.BuildInfo {
	return pkgmodels.BuildInfo{
		Env:            defaultConfig.Env,
		ProjectName:    ProjectName,
		AppName:        AppName,
		AppVersion:     AppVersion,
		CommitHash:     CommitHash,
		BuildTimestamp: BuildTimestamp,                                               // Already in UTC from Makefile
		StartTimestamp: StartTimestamp.UTC().Format(constants.DateTimeLayoutISO8601), // Converted to UTC from local time
		Age:            time.Since(StartTimestamp).String(),
	}
}
