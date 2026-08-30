package smtp

import (
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/spf13/viper"
)

// SmtpListenerConfig holds configuration for the SMTP listener
type SmtpListenerConfig struct {
	Port int    `mapstructure:"port" json:"port"`
	Host string `mapstructure:"host" json:"host"`
}

func NewSmtpListenerConfig(v *viper.Viper) *SmtpListenerConfig {
	smtpListenerDefaults(v)
	cfg := SmtpListenerConfig{}
	if err := v.Unmarshal(&cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed unmarshalling configuration of route %q", v.GetString("Id"))
	}

	return &cfg
}

func smtpListenerDefaults(v *viper.Viper) {
	config.SetDefault(v, "port", 2525)
	config.SetDefault(v, "host", "0.0.0.0")
}
