/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	configOnce sync.Once
	config     Config
)

type Config struct {
	MetricsServerHost   string `mapstructure:"metrics_server_host"`
	MetricsServerScheme string `mapstructure:"metrics_server_scheme"`

	MaxNumOfSamples     int32 `mapstructure:"max_num_of_samples"`
	DefaultNumOfSamples int32 `mapstructure:"default_num_of_samples"`
}

func setConfigDefaults() {
	viper.SetDefault("metrics_server_host", "runai-backend-metrics-service:8080")
	viper.SetDefault("metrics_server_scheme", "http")

	viper.SetDefault("max_num_of_samples", 1000)
	viper.SetDefault("default_num_of_samples", 20)
}

func readConfig() *Config {
	configOnce.Do(func() {
		setConfigDefaults()

		viper.AutomaticEnv()

		err := viper.Unmarshal(&config)
		if err != nil {
			panic(fmt.Sprintf("failed to read metrics client config, error: %v", err))
		}
	})
	return &config
}
