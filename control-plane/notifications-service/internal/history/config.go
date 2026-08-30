package history

import (
	"time"

	"github.com/spf13/viper"
)

type HistoryConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	CleanupConfig `mapstructure:"cleanup"`
}

type CleanupConfig struct {
	Interval         time.Duration `mapstructure:"interval_duration" validate:"gt=0"`
	MaxEventsPerUser int           `mapstructure:"max_events_per_user" validate:"gt=0"`
	MaxAge           time.Duration `mapstructure:"max_age_duration" validate:"gt=0"`
}

func Init(v *viper.Viper) {
	v.SetDefault("enabled", "false")
	v.SetDefault("cleanup.max_events_per_user", 10_000)
	v.SetDefault("cleanup.interval_duration", "1h")
	v.SetDefault("cleanup.max_age_duration", "720h")
}
