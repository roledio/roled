package infra

import (
	"fmt"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/roledio/roled/auth/internal/configs"
	pkgmodels "github.com/roledio/roled/auth/pkg/models"
)

type NewrelicService interface {
	GetApplication() *newrelic.Application
}

type newrelicService struct {
	app *newrelic.Application
}

func NewNewrelicService(defaultConfig *configs.DefaultConfig, buildInfo pkgmodels.BuildInfo) (NewrelicService, error) {
	service := &newrelicService{}
	if !defaultConfig.Newrelic.Enabled {
		return service, nil
	}
	nrapp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(fmt.Sprintf("%s-%s", buildInfo.ProjectName, buildInfo.Env)),
		newrelic.ConfigLicense(defaultConfig.Newrelic.LicenseKey),
		newrelic.ConfigEnabled(defaultConfig.Newrelic.Enabled),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigDatastoreRawQuery(true),
	)
	if err != nil {
		return nil, err
	}
	service.app = nrapp
	return service, nil
}

func (s *newrelicService) GetApplication() *newrelic.Application {
	return s.app
}
