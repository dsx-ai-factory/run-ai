/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

var Schema = "cluster_service"

// Init initializes the configuration system
func Init(secretsFolderPath string) {
	logging.Debugf("initializing config")

	if err := config.SetKbSecretsFolder(secretsFolderPath); err != nil {
		logging.Fatalf("Failed to set KB secrets folder: %v", err)
	}

	if err := registerConfigs(); err != nil {
		logging.Fatalf("Failed to register configs: %v", err)
	}
}

func registerConfigs() error {
	if err := config.Register[*Service](); err != nil {
		return fmt.Errorf("failed to register service configuration: %v", err)
	}

	if err := config.Register[*Server](); err != nil {
		return fmt.Errorf("failed to register server configuration: %v", err)
	}

	if err := config.Register[*Auth](); err != nil {
		return fmt.Errorf("failed to register auth configuration: %v", err)
	}

	if err := config.Register[*Backend](); err != nil {
		return fmt.Errorf("failed to register backend configuration: %v", err)
	}

	if err := config.Register[*App](); err != nil {
		return fmt.Errorf("failed to register app configuration: %v", err)
	}

	if err := config.Register[*Installation](); err != nil {
		return fmt.Errorf("failed to register installation configuration: %v", err)
	}

	if err := config.Register[*CloudFormation](); err != nil {
		return fmt.Errorf("failed to register cloudformation configuration: %v", err)
	}

	if err := config.Register[*Environment](); err != nil {
		return fmt.Errorf("failed to register environment configuration: %v", err)
	}

	if err := config.Register[*Sync](); err != nil {
		return fmt.Errorf("failed to register sync configuration: %v", err)
	}

	if err := config.Register[*AsyncClusterDelete](); err != nil {
		return fmt.Errorf("failed to register async cluster delete configuration: %v", err)
	}

	if err := config.Register[*EditableNodepoolCapabilities](); err != nil {
		return fmt.Errorf("failed to register editable nodepool capabilities configuration: %v", err)
	}

	if err := config.Register[*ApplicationMetricStore](); err != nil {
		return fmt.Errorf("failed to register application metric store configuration: %v", err)
	}

	if err := config.Register[*ImagePullSecret](); err != nil {
		return fmt.Errorf("failed to register image pull secret configuration: %v", err)
	}

	// Register authorization client config
	if err := config.Register[*authorizationclient.Config](); err != nil {
		return fmt.Errorf("failed to register authorization configuration: %v", err)
	}

	return nil
}
