/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/client-go/rest"
)

const NetworkOperatorStepName = "Installing NVIDIA Network Operator"

type NetworkOperatorInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient  k8s.K8sClients
	crdChecker CRDChecker
	version    string
}

func NewNetworkOperatorInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	n := &NetworkOperatorInstallStep{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(NetworkOperatorStepName),
		crdChecker:           &DefaultCRDChecker{},
		k8sClient:            k8sClient,
		version:              version,
	}
	n.SetStatusEvaluator(n.EvaluateStatus)
	return n
}

func (n *NetworkOperatorInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	n.SetStatus(model.Installing, nil)
	if runaiData.GetGPUType() == config.GPUTypeReal {
		logging.Infof("[Prereq] Installing NVIDIA Network Operator %s...", n.version)
		return n.installRealNetworkOperator()
	} else {
		logging.Infof("[Prereq] Skipping Network Operator installation (fake GPU mode)")
		return nil
	}
}

func (n *NetworkOperatorInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if n.ShouldSkipStatusEvaluation() {
		return
	}

	// Fake GPU mode skips network operator entirely
	if runaiData.GetGPUType() == config.GPUTypeFake {
		logging.Infof("[Prereq] Skipping Network Operator status check (fake GPU mode)")
		n.SetStatus(model.Installed, nil)
		return
	}

	logging.Infof("[Prereq] Checking if NVIDIA Network Operator is installed...")
	err := ValidateComponentNamespace(context.Background(), n.k8sClient.GetClient(), "network", "operator", "nvidia-network-operator")
	if err != nil {
		elapsed := n.GetElapsedTime()
		timeout := n.GetTimeout()
		if elapsed > 0 {
			logging.Infof("[Prereq] Network Operator not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Network Operator not found in cluster")
		}
		return
	}

	logging.Infof("[Prereq] NVIDIA Network Operator is installed")
	n.SetStatus(model.Installed, nil)
}

// determineWithConfigAndCRDChecker allows tests to inject a custom CRD checker function
func (n *NetworkOperatorInstallStep) determineWithConfigAndCRDChecker(cfg *rest.Config, crdChecker func(*rest.Config, string) bool) model.InstallPhase {
	crdInstalledOnCluster := crdChecker(cfg, "nicclusterpolicies.nvidia.com")
	if crdInstalledOnCluster {
		return model.Installed
	}
	return model.NotInstalled
}

// installRealNetworkOperator installs the NVIDIA Network Operator using Helm
func (n *NetworkOperatorInstallStep) installRealNetworkOperator() error {
	logging.Infof("[Prereq] Installing NVIDIA Network Operator from NGC repository...")
	kubeVersion := &chartutil.KubeVersion{
		Version: "1.28.15",
		Major:   "1",
		Minor:   "28",
	}
	// Install NVIDIA Network Operator with Helm
	err := helmclient.InstallOrUpgradeReleaseWithKubeVersion(context.Background(),
		"https://helmclient.ngc.nvidia.com/nvidia",
		n.version,
		"network-operator",
		"network-operator",
		"nvidia-network-operator",
		false,
		[]string{},
		kubeVersion,
	)
	if err != nil {
		fullErr := fmt.Errorf("failed to install NVIDIA Network Operator with helm: %v", err)
		logging.Errorf("[Prereq] %v", fullErr)
		n.SetStatus(model.Failed, fullErr)
		return fullErr
	}
	logging.Infof("[Prereq] NVIDIA Network Operator Helm chart deployed, waiting for readiness")
	return nil
}
