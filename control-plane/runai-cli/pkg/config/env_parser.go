package config

import (
	"log"
	"os"
)

type EnvVars struct {
	GinPort     string
	GinMode     string
	MetricsPort string
}

func GetEnvOrDefault(environmentVariableName string, defaultValue string) string {
	env := os.Getenv(environmentVariableName)
	if env == "" {
		log.Printf("environment variable %s was not provided, using default value %v instead", environmentVariableName, defaultValue)
		return defaultValue
	} else {
		return env
	}
}
