package api_exerciser

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type Config struct {
	NUsers        int     `mapstructure:"api_exerciser_users"`
	ThinkTimeSecs float64 `mapstructure:"api_exerciser_think_time"`
	PageSize      int32   `mapstructure:"api_exerciser_page_size"`
	DurationSecs  int     `mapstructure:"api_exerciser_duration_secs"` // 0 = run until signal
}

func (c *Config) SetDefaults() {
	c.NUsers = 5
	c.ThinkTimeSecs = 5.0
	c.PageSize = 50
	c.DurationSecs = 0
}

func (c *Config) validate() error {
	if c.NUsers <= 0 {
		return fmt.Errorf("api_exerciser_users must be > 0, got %d", c.NUsers)
	}
	if c.ThinkTimeSecs < 0 {
		return fmt.Errorf("api_exerciser_think_time must be >= 0, got %g", c.ThinkTimeSecs)
	}
	if c.PageSize <= 0 {
		return fmt.Errorf("api_exerciser_page_size must be > 0, got %d", c.PageSize)
	}
	if c.DurationSecs < 0 {
		return fmt.Errorf("api_exerciser_duration_secs must be >= 0 (0 = run until signal), got %d", c.DurationSecs)
	}
	return nil
}

func ConstructConfig() *Config {
	cfg, err := config.GetR[*Config]()
	if err != nil {
		logging.Fatalf("failed to load api-exerciser config: %v", err)
	}
	if err := cfg.validate(); err != nil {
		logging.Fatalf("invalid api-exerciser config: %v", err)
	}
	return cfg
}
