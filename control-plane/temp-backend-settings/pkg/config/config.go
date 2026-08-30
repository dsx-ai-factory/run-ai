package config

import (
	"sync"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var once sync.Once

func registerConfigurations() {
	if err := config.Register[*ServicesConfig](); err != nil {
		logging.Fatalf("failed to register services configuration: %+v", err)
	}
}

func InitConfig() {
	logging.Infof("initializing configuration")
	once.Do(func() {
		registerConfigurations()
	})
}
