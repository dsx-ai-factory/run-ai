/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName       string `mapstructure:"service_name"`
	CollectorEndpoint string `mapstructure:"tracing_collector_endpoint"`
	UnderlineTracer   string `mapstructure:"tracing_underline_tracer"`
	Enabled           bool   `mapstructure:"tracing_enabled"`
}

func setConfigDefaults() {
	viper.SetDefault("service_name", "")
	viper.SetDefault("tracing_enabled", false)
	viper.SetDefault("tracing_underline_tracer", "otel")
	viper.SetDefault("tracing_collector_endpoint", "otel-collector.monitoring:4317")
}

func readConfig() (*Config, error) {
	setConfigDefaults()

	viper.AutomaticEnv()

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed load tracing config: %v", err)
	}
	return &config, nil
}
