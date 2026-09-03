/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

type daemonSetsTolerationsUpdate struct{}

func newDaemonSetsTolerationsUpgrader() *daemonSetsTolerationsUpdate {
	return &daemonSetsTolerationsUpdate{}
}

func (d *daemonSetsTolerationsUpdate) shouldPerformUpgrade(resource *runaiv1.RunaiConfig) bool {
	if resource.Spec.Global != nil && len(resource.Spec.Global.AdditionalTolerations) > 0 {
		return true
	}
	return false
}

func (d *daemonSetsTolerationsUpdate) updateRunaiConfig(resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error) {
	resource.Spec.Global.DaemonsetsTolerations = resource.Spec.Global.AdditionalTolerations
	resource.Spec.Global.AdditionalTolerations = nil
	return resource, nil
}

func (d *daemonSetsTolerationsUpdate) upgradeHandlerIdentifier() string {
	return "daemonsets-tolerations-update"
}
