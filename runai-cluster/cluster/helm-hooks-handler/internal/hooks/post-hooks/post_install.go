/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	resource_upgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-upgrades"
	controlplanedetailsdownloader "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PostInstallHook struct {
	params *hooks.ClusterHookParams
}

func NewPostInstallHook(params *hooks.ClusterHookParams) *PostInstallHook {
	return &PostInstallHook{
		params: params,
	}
}

func (p *PostInstallHook) GetIdentifier() string {
	return hooks.PostInstallHookIdentifier
}

func (p *PostInstallHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting PostInstall")
	defer logger.Info("Done PostInstall")

	foundInCluster := true
	existingRunaiConfig := &runaiv1.RunaiConfig{}
	err := resources.LocalClusterClient.Get(ctx, types.NamespacedName{
		Name:      runaiv1.DefaultName,
		Namespace: runaiv1.DefaultNamespace,
	}, existingRunaiConfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			foundInCluster = false
		} else {
			return fmt.Errorf("failed to check for an existing runaiconfig: %w", err)
		}
	}

	if foundInCluster {
		logger.Info("runaiconfig already exists, skipping creation")
		return nil
	}

	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}

	installationConfiguration, err := controlplanedetailsdownloader.GetInstallationConfiguration(resources.LocalClusterClient)
	if err != nil {
		return err
	}

	if p.params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade() {
		err = installCPSecrets(resources.RemoteClusterClient, installationConfiguration)
		if err != nil {
			return err
		}
	}

	if p.params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade() ||
		p.params.Development.Headless {
		cluster, err := ra.HandleClusterCR(ctx, resources.LocalClusterClient, p.params.Development.Headless)
		if err != nil {
			return err
		}

		runaiConfig.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				Kind:       runaiv1.ClusterKind,
				Name:       runaiv1.ClusterName,
				UID:        cluster.ObjectMeta.UID,
				APIVersion: runaiv1.GroupVersion.String(),
			},
		}

		if installationConfiguration != nil {
			// Control-plane-assisted installs seed workloads-auth via installCPSecrets above.
			// Headless installs skip that path (ShouldPerformControlPlaneAssistedInstallationOrUpgrade
			// is false), so seed here before Sanitize() strips the client secret from the config.
			if p.params.Development.Headless {
				if err := seedWorkloadsAuthClientSecret(resources.RemoteClusterClient, installationConfiguration); err != nil {
					return err
				}
			}
			installationConfiguration.Internal.Global.ControlPlane.Config.Sanitize()
			runaiConfig.Spec.Internal = &installationConfiguration.Internal
		} else {
			logger.Info("control-plane details secret not found, creating an empty runaiconfig")
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		for attempt := 1; ; attempt++ {
			logger.Info("attempting to create runaiconfig", "attempt", attempt)
			err = commitRunaiconfigToCluster(ctx, resources.LocalClusterClient, runaiConfig, p.params,
				installationConfiguration, false, "")
			if err == nil {
				break
			}
			logger.Error(err, "failed to create runaiconfig", "attempt", attempt, "reason", apierrors.ReasonForError(err), "error", err.Error())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}

		if installationConfiguration.Internal.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOpenShift {
			err = resource_upgrades.NewOcpMonitoringConfigmapUpgrade().Upgrade(ctx, resources)
			if err != nil {
				return fmt.Errorf("failed to upgrade ocp monitoring configmap: %w", err)
			}
		}
	}

	return nil
}
