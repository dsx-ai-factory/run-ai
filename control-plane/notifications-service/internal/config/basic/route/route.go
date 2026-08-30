package route

import (
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter"
	"github.com/spf13/viper"
)

type BasicRouteConfiguration struct {
	RouteId    string                 `mapstructure:"route_id" json:"route_id"`
	ListenerId string                 `mapstructure:"listener_id" json:"listener_id"`
	Enrich     enrich.EnrichingConfig `mapstructure:"enrich" json:"enrich"`
	Filtering  filter.FilteringConfig `mapstructure:"filtering" json:"filtering"`
	NotifyIds  []string               `mapstructure:"notify_ids" json:"notify_ids"`
	Metadata   *Metadata              `mapstructure:"metadata" json:"metadata"`
}

type Metadata map[string]string

func NewBasicConfiguration() *BasicRouteConfiguration {
	return &BasicRouteConfiguration{
		Filtering: filter.FilteringConfig{},
	}
}

// BasicRouteConfigurationProvider - allows access to the BasicRouteConfiguration, should implement app.RouteConfigurationProvider
type BasicRouteConfigurationProvider struct {
	config *BasicRouteConfiguration
	viper  *viper.Viper
}

func (brcp *BasicRouteConfigurationProvider) GetListenerId() string {
	return brcp.config.ListenerId
}

func (brcp *BasicRouteConfigurationProvider) GetRouteId() string {
	return brcp.config.RouteId
}

func NewBasicRouteConfigurationProvider(viper *viper.Viper, routeName string) *BasicRouteConfigurationProvider {
	brcp := &BasicRouteConfigurationProvider{
		viper: viper,
	}
	brcp.init(routeName)
	return brcp
}

func (brcp *BasicRouteConfigurationProvider) init(routeName string) {
	brcp.config = NewBasicConfiguration()

	err := setDefaultConfigValues(brcp.config, brcp.viper)
	if err != nil {
		log.Err(err).Msgf("failed to unmarshal configuration of route %q", routeName)
	}

	if err := brcp.viper.Unmarshal(brcp.config, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed to unmarshal configuration of route %q", routeName)
	}

	brcp.config.RouteId = routeName
}

func (brcp *BasicRouteConfigurationProvider) GetConfig() *BasicRouteConfiguration {
	return brcp.config
}

func setDefaultConfigValues(c *BasicRouteConfiguration, v *viper.Viper) error {
	setRouteDefaults(v)
	setEnrichingDefaults(v)
	err := setFilteringDefaults(&c.Filtering, v)
	return err
}

func setRouteDefaults(v *viper.Viper) {
	v.SetDefault("route.concurrent_limit", 50)
	v.SetDefault("route.cached_events", 1000)
}
