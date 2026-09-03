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
)

type InstallAirgapCPChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller       step.HelmInstaller
	runaiClientProvider step.RunAIAPIClientProvider
}

func NewInstallAirgapCPChart(helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider) step.InstallStep {
	i := &InstallAirgapCPChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install CP Chart (airgap)"),
		helmInstaller:        helmInstaller,
		runaiClientProvider:  runaiClientProvider,
	}
	i.SetStatusEvaluator(i.EvaluateStatus)
	return i
}

func (i *InstallAirgapCPChart) EvaluateStatus(runaiData *model.RunAIInstallData) {
	step.EvaluateCPStatus(i.RunAIInstallStepImpl, i.runaiClientProvider, runaiData)
}

func (i *InstallAirgapCPChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}
	i.SetStatus(model.Installing, nil)

	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	if fqdn == "" {
		i.SetStatus(model.Failed, fmt.Errorf("FQDN is required but not provided"))
		return i.GetError()
	}
	logging.Infof("[Airgap] Using FQDN: %s", fqdn)

	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	extractedDir := runaiData.Config.AirgapExtractedDir
	chartPath := airgap.FindChartFile(extractedDir, step.CPChartName)
	customEnvPath := filepath.Join(extractedDir, airgap.CustomEnvFile)

	values := buildAirgapCPHelmValues(runaiData, fqdn)

	logging.Infof("[Airgap] Installing CP chart from %s with values file %s...", chartPath, customEnvPath)
	err := i.helmInstaller.InstallOrUpgradeFromLocalChart(ctx,
		chartPath, step.CPReleaseName, step.CPNamespace,
		values, []string{customEnvPath})
	if err != nil {
		logging.Errorf("[Airgap] Failed to install Control Plane chart: %v", err)
		i.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Airgap] Control Plane Helm chart deployed")
	return nil
}

func buildAirgapCPHelmValues(runaiData *model.RunAIInstallData, fqdn string) []string {
	values := []string{
		"global.domain=" + fqdn,
		"global.customCA.enabled=true",
	}
	if k8sDist := runaiData.GetInputValue(configinputs.InstallInputKubernetesDistribution); k8sDist != "" {
		values = append(values, "global.config.kubernetesDistribution="+k8sDist)
	}
	return values
}

func (i *InstallAirgapCPChart) GetTimeoutHint() string {
	return "Control Plane did not become accessible within timeout. Check pod status with: kubectl get pods -n runai-backend"
}
