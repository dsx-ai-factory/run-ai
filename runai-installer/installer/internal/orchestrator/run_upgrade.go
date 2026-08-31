/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/upgrade"
)

func RunUpgrade(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps) error {
	logging.Infof("Starting CLI upgrade...")
	step.SetDefaultStepTimeout(upgradeCfg.StepTimeout)

	if err := upgrade.ResolveUpgradeParams(upgradeCfg, deps.HelmValuesReader); err != nil {
		return err
	}

	if deps.ClusterLookup == nil && upgradeCfg.UpgradeCluster {
		deps.ClusterLookup = upgrade.NewClusterLookup(upgradeCfg.FQDN, upgradeCfg.CACertPath, deps.RunAIClientProvider)
	}

	externalEvents := make(chan *flow.FlowStatusReportingMsg, 100)
	flowGroups := buildUpgradeFlowGroups(upgradeCfg, deps, externalEvents)

	if len(flowGroups) == 0 {
		logging.Infof("Nothing to upgrade")
		close(externalEvents)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), upgradeCfg.ExecutionTimeout)
	defer cancel()

	go func() {
		defer close(externalEvents)
		for _, fg := range flowGroups {
			internalEvents := make(chan interface{}, 100)
			fg.Manager.ExecuteFlow(ctx, nil, internalEvents)
			fg.Manager.Wait()
			close(internalEvents)
			if fg.Manager.HasFailed() {
				break
			}
		}
	}()

	var firstFailure string
	for event := range externalEvents {
		switch event.FlowStatus {
		case model.Installing:
			logging.Infof("Upgrading %s...", event.FlowName)
		case model.Installed:
			logging.Infof("Upgraded %s", event.FlowName)
		case model.Failed, model.Timeout:
			logging.Errorf("Failed to upgrade %s", event.FlowName)
			if firstFailure == "" {
				firstFailure = event.FlowName
			}
		}
	}

	if firstFailure != "" {
		return fmt.Errorf("upgrade failed: flow %q did not succeed", firstFailure)
	}
	return nil
}
