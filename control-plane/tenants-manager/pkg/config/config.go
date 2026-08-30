package config

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func Init(secretsFolder string) error {
	logging.Debugf("initializing config")

	if err := config.SetKbSecretsFolder(secretsFolder); err != nil {
		return err
	}

	if err := registerConfigs(); err != nil {
		return err
	}

	return nil
}

func registerConfigs() error {
	if err := config.Register[*Configuration](); err != nil {
		return fmt.Errorf("failed to register configuration: %v", err)
	}

	if err := config.Register[*Tenant](); err != nil {
		return fmt.Errorf("failed to register tenant configuration: %v", err)
	}

	if err := config.Register[*General](); err != nil {
		return fmt.Errorf("failed to register general configuration: %v", err)
	}

	if err := config.Register[*Server](); err != nil {
		return fmt.Errorf("failed to register server configuration: %v", err)
	}

	if err := config.Register[*Auth](); err != nil {
		return fmt.Errorf("failed to register auth configuration: %v", err)
	}

	if err := config.Register[*BffService](); err != nil {
		return fmt.Errorf("failed to register bff service configuration: %v", err)
	}

	return nil
}
