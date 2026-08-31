/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
)

type AdviceRestartControllerConfig struct {
	UseDebugLogLevel     bool   `mapstructure:"USE_DEBUG_LOG_LEVEL"`
	K8sClientConfigQPS   int    `mapstructure:"K8S_CLIENT_CONFIG_QPS"`
	K8sClientConfigBurst int    `mapstructure:"K8S_CLIENT_CONFIG_BURST"`
	Mode                 string `mapstructure:"MODE"`
	RestartLimit         int    `mapstructure:"RESTART_LIMIT"`
	MaxRetriesPerAdvice  int    `mapstructure:"MAX_RETRIES_PER_ADVICE"`
	WebhookPort          int    `mapstructure:"WEBHOOK_PORT"`
	WebhookCertDir       string `mapstructure:"WEBHOOK_CERT_DIR"`
	MetricsBindAddress   string `mapstructure:"METRICS_BIND_ADDRESS"`
}

func (a *AdviceRestartControllerConfig) SetDefaults() {
	a.UseDebugLogLevel = false
	a.K8sClientConfigQPS = 50
	a.K8sClientConfigBurst = 300
	a.Mode = "action"
	a.RestartLimit = 3
	a.MaxRetriesPerAdvice = 3
	a.WebhookPort = 8443
	a.WebhookCertDir = "/etc/webhook/certs"
	a.MetricsBindAddress = ":8383"
}

func Config() AdviceRestartControllerConfig {
	return *rcpconfig.Get[*AdviceRestartControllerConfig]()
}
