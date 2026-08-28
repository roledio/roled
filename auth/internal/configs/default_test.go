package configs

import (
	"testing"

	"github.com/roledio/roled/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_IsEnvProd(t *testing.T) {
	c := &DefaultConfig{Env: constants.EnvProduction}
	assert.True(t, c.IsEnvProd())
	c.Env = "dev"
	assert.False(t, c.IsEnvProd())
}

func TestDefaultConfig_IsEnvLocal(t *testing.T) {
	c := &DefaultConfig{Env: constants.EnvLocal}
	assert.True(t, c.IsEnvLocal())
	c.Env = "prod"
	assert.False(t, c.IsEnvLocal())
}
