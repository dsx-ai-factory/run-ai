package rest

type Config struct{}

func InClusterConfig() (*Config, error) { return &Config{}, nil }
