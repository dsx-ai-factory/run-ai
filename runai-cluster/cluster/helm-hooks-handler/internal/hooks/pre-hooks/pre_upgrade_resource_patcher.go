/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
)

type ResourceUpgradeGroup interface {
	Upgrade(ctx context.Context, clients *clients.Clients) error
}

type resourceUpgrades struct {
	upgrade []ResourceUpgradeGroup
}

func NewResourceGroupUpgrader() *resourceUpgrades {
	var upgradeGroups []ResourceUpgradeGroup
	upgradeGroups = append(upgradeGroups, dynamicupgrades.NewRunaiConfigUpgrade())
	return &resourceUpgrades{upgradeGroups}
}

func (r *resourceUpgrades) Upgrade(ctx context.Context, clients *clients.Clients) error {
	for _, upgradeGroup := range r.upgrade {
		err := upgradeGroup.Upgrade(ctx, clients)
		if err != nil {
			return err
		}
	}
	return nil
}
