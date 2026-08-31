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

type ServerConfig struct {
	ServicePort         int `mapstructure:"service_port"`
	RetryMaxIntervalSec int `mapstructure:"retry_max_interval_sec"`
	RetryMaxElapsedSec  int `mapstructure:"retry_max_elapsed_sec"`
}

func (c *ServerConfig) SetDefaults() {
	c.ServicePort = 8080
	c.RetryMaxIntervalSec = 10
	c.RetryMaxElapsedSec = 60
}

func InitEnvironment() error {
	logging.Debugf("initializing environment variables")
	return registerConfigs()
}

func registerConfigs() error {
	if err := config.Register[*ServerConfig](); err != nil {
		return fmt.Errorf("failed to register server configuration: %v", err)
	}

	return nil
}
