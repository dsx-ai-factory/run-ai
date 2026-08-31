/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ExposePort     int    `mapstructure:"metrics_expose_port"`
	Enabled        bool   `mapstructure:"metrics_enabled"`
	ServiceName    string `mapstructure:"service_name"`
	PushGatewayUrl string `mapstrcture:"metrics_push_gateway_url"`
}

func setConfigDefaults() {
	viper.SetDefault("metrics_expose_port", 9090)
	viper.SetDefault("metrics_enabled", true)
	viper.SetDefault("service_name", "default-service-name")
	viper.SetDefault("metrics_push_gateway_url", "http://prometheus-prometheus-pushgateway.monitoring.svc.cluster.local:9091")
}

func readConfig(c *Config) (*Config, error) {
	if c != nil {
		return c, nil
	}
	setConfigDefaults()

	viper.AutomaticEnv()

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed load metrics config: %v", err)
	}
	return &config, nil
}
