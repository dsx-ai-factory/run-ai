package basic

import "github.com/run-ai/runai/control-plane/notifications-service/internal/app"

type Factory struct {
	configPath string
}

func NewFactory(configPath string) *Factory {
	return &Factory{configPath: configPath}
}

func (f *Factory) CreateConfigurationProvider() app.ConfigurationProvider {
	return NewConfigurationProvider(f.configPath)
}

func (f *Factory) CreateListener() app.ConfigUpdatesListener {
	return NewConfigurationUpdatesListener()
}

func (f *Factory) CreateRoutesRegisters() []app.RoutesRegister {
	return []app.RoutesRegister{NewRoutesRegister()}
}

func (f *Factory) CreateCleaners() []app.Cleaner {
	return []app.Cleaner{}
}
