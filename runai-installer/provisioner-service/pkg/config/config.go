package config

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func Init(secretsFolderPath string) {
	logging.Debugf("initializing config")

	if err := config.SetKbSecretsFolder(secretsFolderPath); err != nil {
		logging.Fatalf("Failed to set secrets folder: %v", err)
	}

	if err := registerConfigs(); err != nil {
		logging.Fatalf("Failed to register configs: %v", err)
	}
}

func registerConfigs() error {
	if err := config.Register[*Server](); err != nil {
		return fmt.Errorf("failed to register server configuration: %v", err)
	}

	if err := config.Register[*TLS](); err != nil {
		return fmt.Errorf("failed to register TLS configuration: %v", err)
	}

	if err := config.Register[*JFrog](); err != nil {
		return fmt.Errorf("failed to register JFrog configuration: %v", err)
	}

	if err := config.Register[*Installation](); err != nil {
		return fmt.Errorf("failed to register Installation configuration: %v", err)
	}

	return nil
}

func ReloadInstallation() {
	if err := config.Register[*Installation](); err != nil {
		logging.Errorf("Failed to reload Installation configuration: %v", err)
	}
}
