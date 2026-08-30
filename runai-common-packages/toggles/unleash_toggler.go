package toggles

import (
	"context"
	"fmt"
	"net/http"
	"time"

	unleash "github.com/Unleash/unleash-client-go/v3"
	ucontext "github.com/Unleash/unleash-client-go/v3/context"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// unleashClienter is the subset of *unleash.Client methods we use, extracted for testability.
type unleashClienter interface {
	IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool)
	Close() error
}

type unleashToggler struct {
	client unleashClienter
}

func newUnleashToggler() (*unleashToggler, error) {
	ut := &unleashToggler{}
	if err := ut.connect(); err != nil {
		return nil, err
	}
	return ut, nil
}

func (ut *unleashToggler) connect() error {
	cfg := config.Get[*Config]()
	opts := []unleash.ConfigOption{
		unleash.WithAppName(cfg.ServiceName),
		unleash.WithUrl(cfg.Url),
		unleash.WithEnvironment(string(cfg.Environment)),
		unleash.WithCustomHeaders(http.Header{"Authorization": {cfg.ApiKey}}),
		unleash.WithRefreshInterval(time.Duration(cfg.RefreshIntervalSeconds) * time.Second),
		unleash.WithListener(&unleashListener{}),
		unleash.WithProjectName(cfg.ProjectName),
	}
	c, err := unleash.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("toggles: failed to create unleash client: %w", err)
	}
	ut.client = c
	log.Infof("toggles: unleash client connected")
	return nil
}

func (ut *unleashToggler) Close() error {
	if ut.client == nil {
		return fmt.Errorf("toggles: cannot close unleash client, not initialized")
	}
	return ut.client.Close()
}

func (ut *unleashToggler) IsEnabled(ctx context.Context, name string) bool {
	if ut.client == nil {
		log.Errorf("toggles: IsEnabled called but unleash client is not initialized, returning false")
		return false
	}
	tenantId, _ := runaictx.ParseTenantId(ctx)
	uctx := ucontext.Context{}
	if tenantId != "" {
		uctx.Properties = map[string]string{
			"tenantId": tenantId,
		}
	}
	return ut.client.IsEnabled(name, unleash.WithContext(uctx))
}
