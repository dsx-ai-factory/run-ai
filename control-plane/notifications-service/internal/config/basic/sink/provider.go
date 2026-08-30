package sink

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/util"
	"github.com/spf13/viper"
)

type ConfigProvider struct {
	SinkConfig map[string]*sinks.SinkConfig // for basic config we cache them all once reading from file.
}

// TODO: maybe can  get the sinkIds from this sub-viper.
func NewSinkConfigProvider(viper *viper.Viper, sinkIds ...string) *ConfigProvider {
	provider := &ConfigProvider{
		SinkConfig: map[string]*sinks.SinkConfig{},
	}
	provider.initSinkConfigs(viper, sinkIds...)
	return provider
}

// ExtractSinkConfigProvider creates a new config provider from an existing one, containing only relevant configs
func ExtractSinkConfigProvider(c *ConfigProvider, sinkIds ...string) *ConfigProvider {
	return &ConfigProvider{SinkConfig: util.FilterMapByKeys(c.SinkConfig, sinkIds...)}
}

func (scp *ConfigProvider) initSinkConfigs(viper *viper.Viper, sinkIds ...string) {
	for _, id := range sinkIds {
		sinkType := viper.Sub(id).GetString("type")
		if config, err := NewSinkConfig(viper.Sub(id), id, sinkType); err != nil {
			log.Err(err)
		} else {
			scp.SinkConfig[id] = config
		}
	}
}

func (scp *ConfigProvider) GetSinks(id ...string) (map[string]*sinks.SinkConfig, error) {
	return util.FilterMapByKeys(scp.SinkConfig, id...), nil
}

func NewSinkConfig(viper *viper.Viper, sinkId string, sinkType string) (*sinks.SinkConfig, error) {
	switch sinkType {
	case sinks.EmailSinkType:
		config, err := sinks.NewEmailSinkConfig(viper, sinkId)
		if err != nil {
			return nil, fmt.Errorf("failed to create email sink: %q", sinkId)
		}
		return &sinks.SinkConfig{Email: config}, nil
	case sinks.SlackSinkType:
		config, err := sinks.NewSlackSinkConfig(viper, sinkId)
		if err != nil {
			return nil, fmt.Errorf("failed to create slack sink: %q", sinkId)
		}
		return &sinks.SinkConfig{Slack: config}, nil
	case sinks.PostgresSinkType:
		config, err := sinks.NewPostgresSinkConfig(viper, sinkId)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres sink: %q", sinkId)
		}
		return &sinks.SinkConfig{Postgres: config}, nil
	case sinks.HistorySinkType:
		return &sinks.SinkConfig{History: sinks.NewHistorySinkConfig(viper, sinkId)}, nil
	default:
		return nil, fmt.Errorf("not supported sink type: %q", viper.GetString("type"))
	}
}
