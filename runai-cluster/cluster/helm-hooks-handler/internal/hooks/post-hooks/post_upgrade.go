/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	resource_upgrades "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-upgrades"
	controlplanedetailsdownloader "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PostUpgradeHook struct {
	params *hooks.ClusterHookParams
}

func NewPostUpgradeHook(params *hooks.ClusterHookParams) *PostUpgradeHook {
	return &PostUpgradeHook{
		params: params,
	}
}

func (p *PostUpgradeHook) GetIdentifier() string {
	return hooks.PostUpgradeHookIdentifier
}

func (p *PostUpgradeHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting PostUpgrade")
	defer logger.Info("Done PostUpgrade")

	if p.params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade() ||
		p.params.Development.Headless {
		_, err := ra.HandleClusterCR(ctx, resources.LocalClusterClient, p.params.Development.Headless)
		if err != nil {
			return err
		}
		// Control-plane-assisted upgrades seed workloads-auth via installCPSecrets below.
		// Headless upgrades skip that path, so seed here (recoverRunaiConfig sanitizes the
		// client secret out of the config it persists to the CRD).
		if p.params.Development.Headless {
			installationConfiguration, err := controlplanedetailsdownloader.GetInstallationConfiguration(resources.LocalClusterClient)
			if err != nil {
				return err
			}
			if err := seedWorkloadsAuthClientSecret(resources.RemoteClusterClient, installationConfiguration); err != nil {
				return err
			}
		}
		err = p.recoverRunaiConfig(ctx, resources.LocalClusterClient)
		if err != nil {
			return err
		}
	}

	if p.params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade() {
		logger.Info("attempting to recover runaiconfig from backup if it exists")
		logger.Info("looking for control-plane details secret")
		installationConfiguration, err := controlplanedetailsdownloader.GetInstallationConfiguration(resources.LocalClusterClient)
		if err != nil {
			return err
		}

		err = installCPSecrets(resources.RemoteClusterClient, installationConfiguration)
		if err != nil {
			return err
		}

		if installationConfiguration.Internal.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOpenShift {
			err = resource_upgrades.NewOcpMonitoringConfigmapUpgrade().Upgrade(ctx, resources)
			if err != nil {
				return fmt.Errorf("failed to upgrade ocp monitoring configmap: %w", err)
			}
		}
	}

	err := p.removeOutdatedOwnershipMetadata(resources.LocalClusterClient)
	if err != nil {
		logger.Error(err, "Got an error trying to remove outdated owner references")
		return err
	}

	err = deleteObsoleteClusterCR(ctx, resources.LocalClusterClient)
	if err != nil {
		return fmt.Errorf("failed to delete obsolete cluster CR: %w", err)
	}

	err = HandleQueueVersionUpgrade(ctx, resources, p.params.QueueValidationIntervalMillis)
	if err != nil {
		return fmt.Errorf("failed to handle queue version upgrade: %w", err)
	}

	if err = p.cleanupLegacyEngineResources(ctx, resources.RemoteClusterClient); err != nil {
		return fmt.Errorf("failed to clean up legacy Engine resources: %w", err)
	}

	if err = p.finalizeResourceManagementMigration(ctx, resources); err != nil {
		return fmt.Errorf("failed to finalize resource-management migration: %w", err)
	}

	if err = dynamicupgrades.NewRunaiConfigEnableAnyworkloadGPUNetworkAcceleration().Upgrade(ctx, resources); err != nil {
		return fmt.Errorf("failed to migrate GPU network acceleration to anyworkload-controller: %w", err)
	}

	return nil
}
