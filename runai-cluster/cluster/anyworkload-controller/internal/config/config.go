/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import "strings"

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

	// SnapshotEnabled reports whether the controller should do anything with
	// NVIDIA Snapshot. The operator sets it from what it detects in the cluster
	// and from its own development gate; it is not a user preference. False
	// means no Snapshot watches are registered and the controller behaves
	// exactly as it does on a cluster without Snapshot.
	SnapshotEnabled bool `mapstructure:"SNAPSHOT_ENABLED"`

	// SnapshotAgentNodeSelector is the node selector on Snapshot's agent
	// DaemonSet, as a comma-separated key=value list. Capture and restore pods
	// can only be scheduled where the agent runs. Empty means no constraint.
	// Kept as the raw string because the config decoder has no string-to-map
	// hook; use AgentNodeSelector for the parsed form.
	SnapshotAgentNodeSelector string `mapstructure:"SNAPSHOT_AGENT_NODE_SELECTOR"`
}

// SetDefaults applies defaults to c.
func (c *Config) SetDefaults() {
	c.RestartEnabled = true
	c.RestartMaxConcurrentReconciles = 10
	c.RestartFrameworkDeletionTimeoutSec = 600
	c.SnapshotEnabled = false
}

// AgentNodeSelector parses SnapshotAgentNodeSelector into the selector to put
// on capture and restore pods. Malformed entries are dropped rather than
// failing: a bad selector would otherwise take the whole controller down over
// something cosmetic.
func (c *Config) AgentNodeSelector() map[string]string {
	if c.SnapshotAgentNodeSelector == "" {
		return nil
	}

	selector := map[string]string{}
	for _, pair := range strings.Split(c.SnapshotAgentNodeSelector, ",") {
		key, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found || key == "" {
			continue
		}
		selector[key] = value
	}
	if len(selector) == 0 {
		return nil
	}
	return selector
}
