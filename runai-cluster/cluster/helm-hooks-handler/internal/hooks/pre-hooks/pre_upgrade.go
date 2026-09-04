/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	resource_actions "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	controlplanedetailsdownloader "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	imagepullsecret "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/image-pull-secret"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PreUpgradeHook struct {
	params   *hooks.ClusterHookParams
	customCA *common.CustomCA
}

func NewPreUpgradeHook(params *hooks.ClusterHookParams, customCA *common.CustomCA) *PreUpgradeHook {
	return &PreUpgradeHook{
		params:   params,
		customCA: customCA,
	}
}

func (p *PreUpgradeHook) GetIdentifier() string {
	return hooks.PreUpgradeHookIdentifier
}

func (p *PreUpgradeHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("Starting %s", p.GetIdentifier()))
	defer logger.Info(fmt.Sprintf("Done %s", p.GetIdentifier()))
	if !p.params.Development.Headless {
		token, err := fetchBearerToken(p.params.ControlPlaneDetails.URL, p.params.ClientId, p.params.ClientSecret)
		if err != nil {
			return fmt.Errorf("failed to authenticate with control plane: %w", err)
		}
		controlPlaneVersion, err := getControlPlaneVersion(p.params.ControlPlaneDetails.URL, token)
		if err != nil {
			return fmt.Errorf("failed to get control plane version: %w", err)
		}

		err = compareVersions(controlPlaneVersion, p.params.ClusterDetails.Version)
		if err != nil {
			return fmt.Errorf("failed to compare control plane version with cluster version: %w", err)
		}
	}

	err := deleteOldPermissionsPreInstallPreUpgrade(resources.RemoteClusterClient)
	if err != nil {
		return fmt.Errorf("failed to delete old permissions: %w", err)
	}
	logger.Info("successfully deleted old permissions in pre-install hook")

	err = addPrivilegedPodSecurityStandardToRunaiNamespace(resources.LocalClusterClient)
	if err != nil {
		return fmt.Errorf("failed to add privileged pod security standard to runai namespace in local cluster: %w", err)
	}

	err = addPrivilegedPodSecurityStandardToRunaiNamespace(resources.RemoteClusterClient)
	if err != nil {
		return fmt.Errorf("failed to add privileged pod security standard to runai namespace in remote cluster: %w", err)
	}

	if p.params.Development.Headless {
		headlessClusterInternalConfig := headlessClusterPlaceholderControlPlaneConfig()
		icd := &controlplanedetailsdownloader.InstallationConfigurationData{
			Internal: *headlessClusterInternalConfig,
		}

		err = controlplanedetailsdownloader.CreateOrReplaceInstallationConfigurationSecret(resources.LocalClusterClient, icd)
		if err != nil {
			return fmt.Errorf("failed to create or update installation configuration secret: %w", err)
		}

		err = p.dynamicRunaiConfigUpgrade(ctx, resources)
		if err != nil {
			return fmt.Errorf("failed to upgrade runai config: %w", err)
		}

		err = resource_actions.BackupRunaiConfig(resources.LocalClusterClient)
		if err != nil {
			return fmt.Errorf("failed to backup runai config: %w", err)
		}
	} else if p.params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade() {
		logger.Info("attempting to download control-plane details to secret")
		icd, err := controlplanedetailsdownloader.DownloadInstallationConfiguration(resources.LocalClusterClient, p.params, p.customCA)
		if err != nil {
			return fmt.Errorf("failed to download control-plane details to local cluster: %w", err)
		}

		err = controlplanedetailsdownloader.CreateOrReplaceInstallationConfigurationSecret(resources.LocalClusterClient, icd)
		if err != nil {
			return fmt.Errorf("failed to create or update installation configuration secret to local cluster: %w", err)
		}

		if resources.RemoteClusterClient != resources.LocalClusterClient {
			// TODO Temporary until all sensitive components are moved to the local cluster
			err = controlplanedetailsdownloader.CreateOrReplaceInstallationConfigurationSecret(resources.RemoteClusterClient, icd)
			if err != nil {
				return fmt.Errorf("failed to create or update installation configuration secret to remote cluster: %w", err)
			}
		}

		err = p.dynamicRunaiConfigUpgrade(ctx, resources)
		if err != nil {
			return fmt.Errorf("failed to upgrade runai config: %w", err)
		}

		err = resource_actions.BackupRunaiConfig(resources.LocalClusterClient)
		if err != nil {
			return fmt.Errorf("failed to backup runai config: %w", err)
		}

		if dockerConfigJSON, ok := icd.Global.ControlPlane.Config.GetDockerConfigJsonOk(); ok {
			err := imagepullsecret.CreateImagePullSecret(resources.LocalClusterClient, runaiv1.DefaultNamespace, hooks.RunAIControlPlaneImagePullSecretName, *dockerConfigJSON)
			if err != nil {
				return fmt.Errorf("failed creating image pull secrets: %v", err)
			}

			err = imagepullsecret.CreateImagePullSecret(resources.RemoteClusterClient, runaiv1.DefaultNamespace, hooks.RunAIControlPlaneImagePullSecretName, *dockerConfigJSON)
			if err != nil {
				return fmt.Errorf("failed creating image pull secrets: %v", err)
			}
		}

		logger.Info("control-plane details successfully downloaded to secret")
	}

	return nil
}
