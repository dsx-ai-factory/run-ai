/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ServiceConfig struct{}

func (c *ServiceConfig) SetDefaults() {}

func InitEnvironment() error {
	logging.Debugf("initializing environment variables")
	return registerConfigs()
}

func registerConfigs() error {
	if err := config.Register[*ServiceConfig](); err != nil {
		return fmt.Errorf("failed to register service configuration: %v", err)
	}

	return nil
}
