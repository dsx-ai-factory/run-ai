/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"path/filepath"

	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
	pkgvalidation "github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/validation"
)

type InstallAirgapClusterChart struct {
	*step.RunAIInstallStepImpl
	installInfoStep     *GetClusterInstallationDetails
	helmInstaller       step.HelmInstaller
	runaiClientProvider step.RunAIAPIClientProvider
}

func NewInstallAirgapClusterChart(installInfoStep *GetClusterInstallationDetails, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider) step.InstallStep {
	i := &InstallAirgapClusterChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install cluster chart (airgap)"),
		installInfoStep:      installInfoStep,
		helmInstaller:        helmInstaller,
		runaiClientProvider:  runaiClientProvider,
	}
	i.SetStatusEvaluator(i.EvaluateStatus)
	return i
}

func (i *InstallAirgapClusterChart) EvaluateStatus(runaiData *model.RunAIInstallData) {
	step.EvaluateClusterStatus(i.RunAIInstallStepImpl, i.runaiClientProvider, i.installInfoStep.createClusterStep.clusterUuid, runaiData)
}

func (i *InstallAirgapClusterChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	i.SetStatus(model.Installing, nil)
	if i.installInfoStep.installInfo == nil {
		err := fmt.Errorf("installInfo is not available")
		logging.Errorf("[Airgap] installInfo is not available: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}

	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	extractedDir := runaiData.Config.AirgapExtractedDir
	chartPath := airgap.FindChartFile(extractedDir, step.ClusterChartName)
	customEnvPath := filepath.Join(extractedDir, airgap.CustomEnvFile)

	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	logging.Infof("[Airgap] Configuring cluster chart for FQDN: %s", fqdn)

	values := []string{
		"controlPlane.url=" + fqdn,
		step.ClusterAuthValue(i.installInfoStep.installInfo),
		"cluster.uid=" + i.installInfoStep.createClusterStep.clusterUuid,
		"cluster.url=" + fqdn,
		"global.customCA.enabled=true",
	}
	if runaiData.GetInputValue(configinputs.InstallInputWorkloadRouting) == pkgvalidation.WorkloadRoutingHost {
		values = append(values, "clusterConfig.global.subdomainSupport=true")
	} else {
		values = append(values, "clusterConfig.global.subdomainSupport=false")
	}

	logging.Infof("[Airgap] Installing cluster chart from %s...", chartPath)
	err := i.helmInstaller.InstallOrUpgradeFromLocalChart(ctx,
		chartPath, step.ClusterReleaseName, step.ClusterNamespace,
		values, []string{customEnvPath})
	if err != nil {
		logging.Errorf("[Airgap] Failed to install cluster chart: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Airgap] Cluster Helm chart deployed")
	return nil
}

func (i *InstallAirgapClusterChart) GetTimeoutHint() string {
	return "Cluster did not report liveness within timeout. Check pod status with: kubectl get pods -n runai"
}
