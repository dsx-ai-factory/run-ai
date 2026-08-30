package config

import (
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
)

type CacheType string

const (
	CTRedis    CacheType = "redis"
	CTNats     CacheType = "nats"
	CTInMemory CacheType = "inmemory"
	CTKeyDB    CacheType = "keydb"
	CTValkey   CacheType = "valkey"
)

type Cache struct {
	Type                      CacheType     `mapstructure:"cache_type" validate:"oneof=redis nats inmemory keydb valkey"`
	Enabled                   *bool         `mapstructure:"cache_enabled"`
	MetricsEnabled            *bool         `mapstructure:"cache_metrics_enabled"`
	Url                       string        `mapstructure:"cache_url" validate:"required_unless=Type inmemory"`
	ReadUrl                   string        `mapstructure:"cache_read_url"`
	User                      string        `mapstructure:"cache_user" config:"secret"`
	Password                  string        `mapstructure:"cache_password" config:"secret"`
	TtlSeconds                int           `mapstructure:"cache_ttl_seconds" validate:"required,gt=0"`
	MaxStoreSizeMB            int           `mapstructure:"cache_max_store_size_mb" validate:"required,min=20,max=4000"`
	MaxRecordSizeKB           int           `mapstructure:"cache_max_record_size_kb" validate:"required,min=1,max=30000"`
	ServiceName               string        `mapstructure:"service_name" validate:"required"`
	ConnectRetries            int           `mapstructure:"cache_connect_retries" validate:"gte=0"`
	InitRetryIntervalDuration time.Duration `mapstructure:"cache_init_retry_interval_duration"`
}

func (c *Cache) SetDefaults() {
	c.Enabled = ptr(true)
	c.MetricsEnabled = ptr(true)
	c.Type = "inmemory"
	c.TtlSeconds = 3600
	c.MaxStoreSizeMB = 100
	c.MaxRecordSizeKB = 1
	c.ConnectRetries = 10
	c.InitRetryIntervalDuration = 2 * time.Second
}

func InitConfig() error {
	var err error
	if err = config.SetKbSecretsFolder(""); err != nil {
		return fmt.Errorf("failed to set secrets folder: %w", err)
	}
	if err = config.Register[*Cache](); err != nil {
		return fmt.Errorf("failed to register cache config: %w", err)
	}

	return nil
}

func ptr[T any](x T) *T {
	return &x
}
