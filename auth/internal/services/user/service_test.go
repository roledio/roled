package user

import (
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
)

func newDefaultConfig() *configs.DefaultConfig {
	cfg := &configs.DefaultConfig{
		BaseURL: "http://localhost",
	}
	cfg.Upload.Driver = constants.UploadDriverLocal
	return cfg
}
