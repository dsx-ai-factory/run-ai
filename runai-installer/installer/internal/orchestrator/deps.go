/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type InstallDeps struct {
	K8sClient           k8s.K8sClients
	HelmInstaller       step.HelmInstaller
	RunAIClientProvider step.RunAIAPIClientProvider
}

func ProductionInstallDeps(k8sClient k8s.K8sClients, controlPlaneVersion string) InstallDeps {
	deps := InstallDeps{
		K8sClient:     k8sClient,
		HelmInstaller: step.NewHelmInstaller(),
	}
	if k8sClient != nil {
		deps.RunAIClientProvider = step.NewRunAIAPIClientProvider(apiclient.NewObjectGetter(k8sClient.GetClient()), controlPlaneVersion)
	}
	return deps
}
