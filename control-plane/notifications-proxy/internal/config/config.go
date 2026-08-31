/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func Init() {
	err := config.Register[*ServiceConfig]()
	if err != nil {
		logging.Fatalf("Failed to initialize service config: %v", err)
	}

	err = config.Register[*slack.SlackConfig]()
	if err != nil {
		logging.Fatalf("Failed to initialize slack config: %v", err)
	}
}

type ServiceConfig struct {
	NotifyConfigUrl     string `mapstructure:"notify_config_url"`
	RunaiEnvDomain      string `mapstructure:"runai_env_domain" validate:"required"`
	MultiTenancyEnabled *bool  `mapstructure:"multi_tenancy_enabled"`
}

func (c *ServiceConfig) SetDefaults() {
	c.NotifyConfigUrl = "http://runai-backend-notifications-service:8080"
}
