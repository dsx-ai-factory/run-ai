/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	"context"
	"fmt"

	controlplanedetailsdownloader "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type runaiConfigDynamicSchemaUpgrader interface {
	updateRunaiConfig(resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error)
	shouldPerformUpgrade(resource *runaiv1.RunaiConfig) bool
	upgradeHandlerIdentifier() string
}

type runaiConfigUpgrade struct {
	specUpgraders []runaiConfigDynamicSchemaUpgrader
}

func NewRunaiConfigUpgrade(cpInstallationData *controlplanedetailsdownloader.InstallationConfigurationData, clusterCr *runaiv1.Cluster, previousVersion string) *runaiConfigUpgrade {
	specUpgraders := commonSpecUpgraders(cpInstallationData, previousVersion)
	specUpgraders = append(specUpgraders, newOwnerReferenceUpgrader(clusterCr))
	return &runaiConfigUpgrade{specUpgraders}
}

func NewHeadlessRunaiConfigUpgrade(cpInstallationData *controlplanedetailsdownloader.InstallationConfigurationData, previousVersion string) *runaiConfigUpgrade {
	specUpgraders := commonSpecUpgraders(cpInstallationData, previousVersion)
	return &runaiConfigUpgrade{specUpgraders}
}

func (r *runaiConfigUpgrade) UpgradeRunaiConfigSpec(ctx context.Context, resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error) {
	var err error
	for _, upgrader := range r.specUpgraders {
		resource, err = r.upgradeSpec(ctx, resource, upgrader)
		if err != nil {
			return nil, err
		}
	}
	return resource, nil
}

func (r *runaiConfigUpgrade) upgradeSpec(ctx context.Context, resource *runaiv1.RunaiConfig, upgrader runaiConfigDynamicSchemaUpgrader) (*runaiv1.RunaiConfig, error) {
	logger := log.FromContext(ctx)
	shouldUpgrade := upgrader.shouldPerformUpgrade(resource)
	if !shouldUpgrade {
		logger.Info(fmt.Sprintf("No runaiconfig upgrade needed for handler:[%s]", upgrader.upgradeHandlerIdentifier()))
		return resource, nil
	}
	resource, err := upgrader.updateRunaiConfig(resource)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Completed runaiconfig upgrade , for handler:[%s]", upgrader.upgradeHandlerIdentifier()))
	return resource, nil

}

func commonSpecUpgraders(cpInstallationData *controlplanedetailsdownloader.InstallationConfigurationData, previousVersion string) []runaiConfigDynamicSchemaUpgrader {
	return []runaiConfigDynamicSchemaUpgrader{
		newPrometheusUpgrader(cpInstallationData),
		newDaemonSetsTolerationsUpgrader(),
		newSubdomainSupportUpgrader(previousVersion),
	}
}
