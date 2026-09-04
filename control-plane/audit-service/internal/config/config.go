/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	defaultBatchIntervalSec = 10
	defaultBatchSize        = 10
)

func Init() {
	if err := config.Register[*ServiceConfig](); err != nil {
		logging.Fatalf("Failed to initialize config: %v", err)
	}
	if err := config.Register[*BatcherConfig](); err != nil {
		logging.Fatalf("Failed to initialize batcher config: %v", err)
	}
}

type ServiceConfig struct {
	CleanDbEnabled     bool `mapstructure:"audit_record_cleaner_enabled"`
	CleanIntervalHours int  `mapstructure:"audit_record_cleaner_interval_hours"`
	CleanAuditTtlDays  int  `mapstructure:"audit_record_ttl_days"`
}

func (c *ServiceConfig) SetDefaults() {
	c.CleanDbEnabled = true
	c.CleanIntervalHours = 24
	c.CleanAuditTtlDays = 90
}

type BatcherConfig struct {
	MaxSize     int `mapstructure:"batch_max_size"`
	IntervalSec int `mapstructure:"batch_interval_sec"`
}

func (b *BatcherConfig) SetDefaults() {
	b.MaxSize = defaultBatchSize
	b.IntervalSec = defaultBatchIntervalSec
}
