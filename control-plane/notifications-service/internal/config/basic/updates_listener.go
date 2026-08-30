package basic

import "context"

type ConfigurationUpdatesListener struct{}

func NewConfigurationUpdatesListener() *ConfigurationUpdatesListener {
	return &ConfigurationUpdatesListener{}
}

func (b *ConfigurationUpdatesListener) Listen(context.Context, func(listenerId string) error) {
}
