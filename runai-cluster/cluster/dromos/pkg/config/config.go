package config

// DromosConfig is the controller-wide configuration for the dromos library.
type DromosConfig struct {
	UseDebugLogLevel bool `mapstructure:"USE_DEBUG_LOG_LEVEL"`
	MetricsPort      int  `mapstructure:"METRICS_PORT"`

	K8sClientConfigQPS   int    `mapstructure:"K8S_CLIENT_CONFIG_QPS"`
	K8sClientConfigBurst int    `mapstructure:"K8S_CLIENT_CONFIG_BURST"`
	WebhookPort          int    `mapstructure:"WEBHOOK_PORT"`
	CertDir              string `mapstructure:"WEBHOOK_CERT_DIR"`
	MaxMutationRounds    int    `mapstructure:"MAX_MUTATION_ROUNDS"`
}

// Default returns a DromosConfig with sensible defaults.
func Default() *DromosConfig {
	return &DromosConfig{
		MetricsPort:          9400,
		K8sClientConfigQPS:   50,
		K8sClientConfigBurst: 300,
		WebhookPort:          9443,
		CertDir:              "/etc/webhook/certs",
		MaxMutationRounds:    5,
	}
}

// SetDefaults applies defaults to c.
func (c *DromosConfig) SetDefaults() {
	*c = *Default()
}
