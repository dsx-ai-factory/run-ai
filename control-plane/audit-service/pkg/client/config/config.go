/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

const (
	MaxMetadataProperties = 10
	ClientIdMetadataKey   = "clientId"
)

func InitConfig(c *Config) error {
	if c != nil {
		return config.Set(c)
	}

	if err := config.Register[*Config](); err != nil {
		return fmt.Errorf("failed to register pubsub config: %w", err)
	}

	return nil
}

type Config struct {
	Enabled                      bool   `mapstructure:"audit_enabled"`
	EnablePrintAuditLogsToStdout bool   `mapstructure:"audit_enable_print_to_stdout"`
	AuditServiceUrl              string `mapstructure:"audit_service_url"`
}

func (c *Config) SetDefaults() {
	c.Enabled = true
	c.EnablePrintAuditLogsToStdout = false
	c.AuditServiceUrl = "http://runai-backend-audit-service:8080"
}
