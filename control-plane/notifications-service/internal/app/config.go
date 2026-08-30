package app

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
)

type AppConfig struct {
	LogLevel string `mapstructure:"log_level"`
	APIHost  string `mapstructure:"api_host"`
	APIPort  string `mapstructure:"api_port"`
}

type AppConfigProvider interface {
	GetAppConfig() *AppConfig
}
type ConfigurationProvider interface {
	GetAppConfig() *AppConfig
	GetHistoryConfig() *history.HistoryConfig
	GetListeningProviders() map[string]listen.ListeningConfigProvider
	GetListeningProvider(id string) listen.ListeningConfigProvider

	GetName() string

	route.RouterConfigProvider
}
type ConfigUpdatesListener interface {
	Listen(ctx context.Context, handleListenerChange func(listenerId string) error)
}

type Cleaner interface {
	Schedule(ctx context.Context)
}

type ConfigFactory interface {
	CreateConfigurationProvider() ConfigurationProvider
	CreateListener() ConfigUpdatesListener
	CreateRoutesRegisters() []RoutesRegister
	CreateCleaners() []Cleaner
}
