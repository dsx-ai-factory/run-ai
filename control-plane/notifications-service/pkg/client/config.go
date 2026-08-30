package client

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	DEFAULT_NOTIFY_BATCH_SIZE         = int(20)
	DEFAULT_NOTIFY_BATCH_INTERVAL_SEC = int(5)
	DEFAULT_NOTIFY_SERVICE_URL        = "http://notifications-service:5000"
	DEFAULT_CONFIGURATIONS_URL        = "http://notifications-service:8080"
	DEFAULT_NOTIFY_RETRY_THRESHOLD    = 1
	DEFAULT_NOTIFY_RETRY_INTERVAL_SEC = int(5)
	DEFAULT_CACHE_MAX_SIZE            = 100
	DEFAULT_CACHE_TTL_SEC             = time.Hour
)

type loadConfigDefaults func(config *NotifyClientConfig) (*NotifyClientConfig, error)

type NotifyClientConfig struct {
	BatchSize        int    `mapstructure:"batch_size" json:"batch_size,omitempty" validate:"min=1"`
	BatchIntervalSec int    `mapstructure:"batch_interval_sec" json:"batch_interval_sec,omitempty" validate:"min=0"`
	NotifyUrl        string `mapstructure:"notify_url" json:"notify_url,omitempty"`
	ConfigUrl        string `mapstructure:"config_url" json:"config_url,omitempty"`
	RetryThreshold   int    `mapstructure:"retry_threshold" json:"retry_threshold,omitempty" validate:"min=0"`
	RetryIntervalSec int    `mapstructure:"retry_interval_sec" json:"retry_interval_sec,omitempty" validate:"min=1"`
	CacheSize        int    `mapstructure:"cache_size" json:"cache_size,omitempty" validate:"min=1"`
	CacheTTLSec      int    `mapstructure:"cache_ttl_sec" json:"cache_ttl_sec,omitempty" validate:"min=0"`
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("batch_size", DEFAULT_NOTIFY_BATCH_SIZE)
	v.SetDefault("batch_interval_sec", DEFAULT_NOTIFY_BATCH_INTERVAL_SEC)
	v.SetDefault("notify_url", DEFAULT_NOTIFY_SERVICE_URL)
	v.SetDefault("config_url", DEFAULT_CONFIGURATIONS_URL)
	v.SetDefault("retry_threshold", DEFAULT_NOTIFY_RETRY_THRESHOLD)
	v.SetDefault("retry_interval_sec", DEFAULT_NOTIFY_RETRY_INTERVAL_SEC)
	v.SetDefault("cache_size", DEFAULT_CACHE_MAX_SIZE)
	v.SetDefault("cache_ttl_sec", DEFAULT_CACHE_TTL_SEC)
}

func LoadDefaultsToConfig(config *NotifyClientConfig) (*NotifyClientConfig, error) {
	if config == nil {
		config = &NotifyClientConfig{}
	}

	v := viper.New()
	confStr, err := json.Marshal(*config)
	if err != nil {
		return nil, err
	}

	v.SetConfigType("json")
	err = v.ReadConfig(strings.NewReader(string(confStr)))
	if err != nil {
		return nil, err
	}

	v.AutomaticEnv()
	setDefaults(v)
	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(*config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getHostAndScheme(fullUrl string) (string, string) {
	pUrl, err := url.Parse(fullUrl)
	if err != nil {
		return "", ""
	}

	return pUrl.Host, pUrl.Scheme
}
