package newrelic

import (
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Config struct {
	Enabled    bool
	LicenseKey string
	AppName    string
}

type Service interface {
	GetApplication() *newrelic.Application
}

type service struct {
	app *newrelic.Application
}

func NewService(cfg Config) (Service, error) {
	service := &service{}
	if !cfg.Enabled {
		return service, nil
	}
	nrapp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.LicenseKey),
		newrelic.ConfigEnabled(cfg.Enabled),
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

func (s *service) GetApplication() *newrelic.Application {
	return s.app
}
