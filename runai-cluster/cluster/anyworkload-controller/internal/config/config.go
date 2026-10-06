/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

// Config is the anyworkload-controller's controller-wide configuration. Every
// knob owned by the restart machinery is prefixed RESTART_.
type Config struct {
	// RestartEnabled is toggle for the restart feature
	RestartEnabled bool `mapstructure:"RESTART_ENABLED"`

	// RestartMaxConcurrentReconciles number of restart concurrent reconciles
	// --restart-handoff-max-concurrent-reconciles.
	RestartMaxConcurrentReconciles int `mapstructure:"RESTART_MAX_CONCURRENT_RECONCILES"`

	// RestartFrameworkDeletionTimeoutSec is how long a restart waits for the child workload
	// object to be fully deleted before failing the RestartRequest.
	RestartFrameworkDeletionTimeoutSec int `mapstructure:"RESTART_FRAMEWORK_DELETION_TIMEOUT_SEC"`

	// RestartAllowedServiceAccounts are exact usernames allowed to write spec.managedBy.
	// The operator includes this controller's service account and any configured extra identities.
	RestartAllowedServiceAccounts []string `mapstructure:"RESTART_ALLOWED_SERVICE_ACCOUNTS"`
}

// SetDefaults applies defaults to c.
func (c *Config) SetDefaults() {
	c.RestartEnabled = true
	c.RestartMaxConcurrentReconciles = 10
	c.RestartFrameworkDeletionTimeoutSec = 600
}
