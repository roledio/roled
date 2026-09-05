package models

import (
	"errors"

	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/pkg/constants"
	pkgerrors "github.com/roledio/roled/auth/pkg/errors"
	"github.com/roledio/roled/auth/pkg/models"
)

type TemplateData struct {
	Data      any
	Flash     any
	Error     pkgerrors.CustomError
	IsError   bool
	Config    *configs.DefaultConfig
	BuildInfo models.BuildInfo
	Map       map[string]any
}

func (t *TemplateData) WithError(err error) *TemplateData {
	var ce pkgerrors.CustomError
	if !errors.As(err, &ce) {
		// This error is not of type errors.CustomError, it is an unexpected error
		ce = pkgerrors.ErrSystemError.WithError(err)
	}
	t.Error = ce
	t.IsError = true
	return t
}

func (t *TemplateData) IsEnvProd() bool {
	return t.BuildInfo.Env == constants.EnvProduction
}
