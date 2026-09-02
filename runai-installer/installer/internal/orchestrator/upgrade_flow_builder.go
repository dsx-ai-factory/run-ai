/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/flow"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade/steps"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

func buildUpgradeFlowGroups(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
	if upgradeCfg.IsAirgap() {
		return buildAirgapUpgradeFlowGroups(upgradeCfg, deps, externalEvents)
	}
	return buildStandardUpgradeFlowGroups(upgradeCfg, deps, externalEvents)
}

func buildStandardUpgradeFlowGroups(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
	var flowGroups []*FlowGroup

	if upgradeCfg.UpgradeControlPlane {
		cpFlows := getControlPlaneUpgradeFlows(upgradeCfg, deps)
		if len(cpFlows) > 0 {
			flowGroups = append(flowGroups, &FlowGroup{
				Flows:   cpFlows,
				Manager: flow.NewFlowManagerWithExternalEvents(cpFlows, externalEvents, false),
			})
		}
	}

	if upgradeCfg.UpgradeCluster {
		clusterFlows := getClusterUpgradeFlows(upgradeCfg, deps)
		if len(clusterFlows) > 0 {
			flowGroups = append(flowGroups, &FlowGroup{
				Flows:   clusterFlows,
				Manager: flow.NewFlowManagerWithExternalEvents(clusterFlows, externalEvents, false),
			})
		}
	}

	if len(flowGroups) == 0 {
		logging.Infof("No upgrade flows configured")
	}

	return flowGroups
}

func buildAirgapUpgradeFlowGroups(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
	var flowGroups []*FlowGroup

	extractStep := steps.NewExtractAirgapPackage(upgradeCfg)
	pushStep := steps.NewPushAirgapImages(upgradeCfg)
	prepFlows := []*flow.InstallFlow{
		flow.NewInstallFlow("Airgap Prepare", []step.InstallStep{extractStep, pushStep}, nil),
	}
	flowGroups = append(flowGroups, &FlowGroup{
		Flows:   prepFlows,
		Manager: flow.NewFlowManagerWithExternalEvents(prepFlows, externalEvents, false),
	})

	if upgradeCfg.UpgradeControlPlane {
		cpStep := steps.NewUpgradeAirgapCPChart(upgradeCfg, deps.HelmInstaller)
		cpFlows := []*flow.InstallFlow{
			flow.NewInstallFlow("Upgrade Control Plane", []step.InstallStep{cpStep}, nil),
		}
		flowGroups = append(flowGroups, &FlowGroup{
			Flows:   cpFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(cpFlows, externalEvents, false),
		})
	}

	if upgradeCfg.UpgradeCluster {
		getInfo := steps.NewGetUpgradeInstallInfo(upgradeCfg, deps.ClusterLookup)
		upgradeChart := steps.NewUpgradeAirgapClusterChart(upgradeCfg, getInfo, deps.HelmInstaller)
		clusterFlows := []*flow.InstallFlow{
			flow.NewInstallFlow("Upgrade Cluster", []step.InstallStep{getInfo, upgradeChart}, nil),
		}
		flowGroups = append(flowGroups, &FlowGroup{
			Flows:   clusterFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(clusterFlows, externalEvents, false),
		})
	}

	return flowGroups
}

func getControlPlaneUpgradeFlows(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps) []*flow.InstallFlow {
	logging.Infof("Building control plane upgrade flows (target version: %s)", upgradeCfg.ControlPlaneVersion)
	cpStep := steps.NewUpgradeCPChart(upgradeCfg, deps.HelmInstaller)
	return []*flow.InstallFlow{
		flow.NewInstallFlow("Upgrade Control Plane", []step.InstallStep{cpStep}, nil),
	}
}

func getClusterUpgradeFlows(upgradeCfg *config.UpgradeConfig, deps upgrade.Deps) []*flow.InstallFlow {
	logging.Infof("Building cluster upgrade flows (target version: %s, cluster UUID: %s)", upgradeCfg.ClusterVersion, upgradeCfg.ClusterUUID)
	getInfo := steps.NewGetUpgradeInstallInfo(upgradeCfg, deps.ClusterLookup)
	upgradeChart := steps.NewUpgradeClusterChart(upgradeCfg, getInfo, deps.HelmInstaller)

	return []*flow.InstallFlow{
		flow.NewInstallFlow("Upgrade Cluster", []step.InstallStep{getInfo, upgradeChart}, nil),
	}
}
