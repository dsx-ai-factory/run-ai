/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import (
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type FeatureFlagSystemEnv = string

const (
	FFSEProd FeatureFlagSystemEnv = "production"
	FFSEDev  FeatureFlagSystemEnv = "development"
)

type Config struct {
	Enabled                *bool                `mapstructure:"feature_flag_system_enabled"`
	Url                    string               `mapstructure:"feature_flag_system_url" validate:"required_if=Enabled true"`
	Environment            FeatureFlagSystemEnv `mapstructure:"feature_flag_system_env" validate:"required_if=Enabled true"`
	ApiKey                 string               `mapstructure:"feature_flag_system_api_key" config:"secret"`
	RefreshIntervalSeconds int                  `mapstructure:"feature_flag_system_refresh_interval_seconds" validate:"gte=1"`
	ServiceName            string               `mapstructure:"service_name" validate:"required_if=Enabled true"`
	ProjectName            string               `mapstructure:"feature_flag_system_project_name" validate:"required_if=Enabled true"`
}

func (c *Config) SetDefaults() {
	c.Enabled = ptr.Bool(true)
	c.RefreshIntervalSeconds = 15
	c.ProjectName = "Runai"
}

func registerConfig() error {
	if err := config.SetKbSecretsFolder(""); err != nil {
		return fmt.Errorf("toggles: failed to set secrets folder: %w", err)
	}
	if err := config.Register[*Config](); err != nil {
		return fmt.Errorf("toggles: failed to register config: %w", err)
	}
	return nil
}
