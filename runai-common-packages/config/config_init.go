package config

type Config struct {
	KeyDelimiter string `validate:"required,excludesall= "`
}

func getConfig() *Config {
	return &Config{
		KeyDelimiter: getEnvOrDefault(envConfigKeyDelimiter, defaultKeyDelimiter),
	}
}
