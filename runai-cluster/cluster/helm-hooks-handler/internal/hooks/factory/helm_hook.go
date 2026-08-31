/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"errors"
	"log"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	post_hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/post-hooks"
	pre_hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/pre-hooks"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/installation/util"
	v1 "k8s.io/api/core/v1"
)

func HookFactory(opts *util.Options) (hooks.ClusterHook, error) {
	params, err := getClusterHookParams(opts)
	if err != nil {
		log.Fatalf("failed to get cluster hook params: %v", err)
	}
	if opts.ValidateClusterConfig {
		return pre_hooks.NewValidateClusterConfigHook(params), nil
	}
	if opts.PreInstall {
		return pre_hooks.NewPreInstallHook(params, getCustomCA(opts)), nil
	}
	if opts.PostInstall {
		return post_hooks.NewPostInstallHook(params), nil
	}
	if opts.PreUpgrade {
		return pre_hooks.NewPreUpgradeHook(params, getCustomCA(opts)), nil
	}
	if opts.PreCrdUpgrade {
		return pre_hooks.NewPreCrdUpgradeHook(params), nil
	}
	if opts.PreResourceManagementUpgrade {
		return pre_hooks.NewPreResourceManagementUpgradeHook(params), nil
	}
	if opts.PostUpgrade {
		return post_hooks.NewPostUpgradeHook(params), nil
	}
	if opts.PreDelete {
		return pre_hooks.NewPreDeleteHook(), nil
	}
	return nil, errors.New("no hook action specified")
}

func getCustomCA(opts *util.Options) *common.CustomCA {
	return &common.CustomCA{
		Enabled: &opts.CustomCaEnabled,
		Secret: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: opts.CustomCaSecretName,
			},
			Key:      opts.CustomCaSecretKey,
			Optional: &opts.CustomCaSecretOptional,
		},
	}
}

func getClusterHookParams(opts *util.Options) (*hooks.ClusterHookParams, error) {
	return &hooks.ClusterHookParams{
		ControlPlaneDetails: clients.ControlPlaneDetails{
			URL:            opts.ControlPlaneURL,
			ClientId:       opts.ControlPlaneClientID,
			ClientSecret:   opts.ControlPlaneClientSecret.String(),
			BootstrapToken: opts.ControlPlaneBootstrapToken.String(),
		},
		ClusterDetails: hooks.ClusterDetails{
			UID:     opts.ClusterUID,
			URL:     opts.ClusterURL,
			Version: opts.ClusterVersion,
		},
		ImageRegistry:              opts.ImageRegistry,
		AdditionalImagePullSecrets: opts.AdditionalImagePullSecrets,
		Affinity:                   opts.CMConfigParams.Affinity,
		Tolerations:                opts.CMConfigParams.Tolerations,
		ClusterConfig:              opts.CMConfigParams.ClusterConfig,
		Development: hooks.DevelopmentParams{
			Headless:             opts.Development.Headless,
			SkipDependencyChecks: opts.Development.SkipDependencyChecks,
		},
		NMC: hooks.NMCParams{
			Enabled:                  opts.EnableNMC,
			RemoteServerAPIServerURL: opts.NMCRemoteClusterURL,
			ServiceAccountName:       opts.NMCServiceAccountName,
			AllowedServiceAccounts:   opts.NMCAllowedServiceAccounts,
			MgmtPrometheusIngress:    opts.CMConfigParams.MgmtPrometheusIngress,
		},
		DisableInfrastructureMetricsRemoteWrite:     opts.DisableInfrastructureMetricsRemoteWrite,
		QueueValidationIntervalMillis:               opts.QueueValidationIntervalMillis,
		OperandDownIntervalMillis:                   opts.OperandDownIntervalMillis,
		ResourceManagementMigrationEnabled:          opts.ResourceManagementMigrationEnabled,
		ResourceManagementMigrationCleanup:          opts.ResourceManagementMigrationCleanup,
		ResourceManagementMigrationNodePoolsEnabled: opts.ResourceManagementMigrationNodePoolsEnabled,
		ResourceManagementMigrationNodePoolsCleanup: opts.ResourceManagementMigrationNodePoolsCleanup,
		EngineToKAIMigrationEnabled:                 opts.EngineToKAIMigrationEnabled,
		EnableRbacEscalation:                        opts.EnableRbacEscalation,
		FipsMode:                                    opts.FipsMode,
	}, nil
}
