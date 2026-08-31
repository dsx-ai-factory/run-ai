/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ConfigData struct {
	Namespace        string `mapstructure:"namespace"`
	ConfigFilePath   string `mapstructure:"config_file_path"`
	FullnameOverride string `mapstructure:"fullname_override"`
	LogLevel         string `mapstructure:"log_level"`
	DryRun           bool   `mapstructure:"dry_run"`
}

func (c *ConfigData) SetDefaults() {
	c.Namespace = "runai-backend"
	c.ConfigFilePath = "/etc/secret-generator/secrets.json"
	c.FullnameOverride = "runai-backend"
	c.LogLevel = "Info"
	c.DryRun = false
}

var Config *ConfigData

func Init() *ConfigData {
	logging.Debugf("initializing secret-generator config")
	conf, err := config.GetR[*ConfigData]()
	if err != nil {
		logging.Fatalf("failed to initialize config: %v", err)
	}
	Config = conf
	return conf
}
