package logging

import "github.com/run-ai/runai/runai-common-packages/config"

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Pretty   bool   `mapstructure:"log_pretty"`
}

func (c *Config) SetDefaults() {
	c.LogLevel = "Info"
	c.Pretty = false
}

func readConfig() *Config {
	config, err := config.GetR[*Config]()
	if err != nil {
		Fatalf("failed to initialize logger, error: %v", err)
	}
	return config
}
