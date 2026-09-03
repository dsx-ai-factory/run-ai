/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

const (
	RunaiRegCredsSecret       = "runai-reg-creds"
	RunaiBackendTlsSecretName = "runai-backend-tls"
	RunaiBackendCAChainSecret = "runai-ca-cert"
)

type InstallCPChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller       step.HelmInstaller
	runaiClientProvider step.RunAIAPIClientProvider
}

func NewInstallCPChart(helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider) step.InstallStep {
	i := &InstallCPChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Install CP Chart"),
		helmInstaller:        helmInstaller,
		runaiClientProvider:  runaiClientProvider,
	}
	i.SetStatusEvaluator(i.EvaluateStatus)
	return i
}

func (i *InstallCPChart) EvaluateStatus(runaiData *model.RunAIInstallData) {
	step.EvaluateCPStatus(i.RunAIInstallStepImpl, i.runaiClientProvider, runaiData)
}

func (r *InstallCPChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := r.CheckContextCancellation(ctx); err != nil {
		return err
	}

	// Update status to installing and start timer
	r.SetStatus(model.Installing, nil)

	// Get the FQDN from the input data
	fqdn := runaiData.GetInputValue(configinputs.InstallInputFQDN)
	if fqdn == "" {
		logging.Errorf("[RunAI] FQDN is required but not provided")
		r.SetStatus(model.Failed, fmt.Errorf("FQDN is required but not provided"))
		return r.GetError()
	}
	logging.Infof("[RunAI] Using FQDN: %s", fqdn)

	// Check for context cancellation before helm operation
	if err := r.CheckContextCancellation(ctx); err != nil {
		return err
	}

	values := r.buildHelmValues(runaiData, fqdn)
	version := runaiData.Config.CpVersion
	repo := r.getChartRepo(runaiData.Config.Env)
	logging.Infof("[RunAI] Installing Control Plane chart version %s from %s...", version, repo)
	err := r.helmInstaller.InstallOrUpgradeRelease(ctx,
		repo,
		version, step.CPReleaseName, step.CPChartName, step.CPNamespace, false, values)
	if err != nil {
		logging.Errorf("[RunAI] Failed to install Control Plane chart: %v", err)
		r.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[RunAI] Control Plane Helm chart deployed, waiting for readiness")
	return nil
}

func (i *InstallCPChart) getChartRepo(env config.RunaiEnvType) string {
	return step.CPChartRepo(env)
}

func (r *InstallCPChart) buildHelmValues(runaiData *model.RunAIInstallData, fqdn string) []string {
	values := []string{"global.domain=" + fqdn}

	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		values = append(values, "global.customCA.enabled=true")
		logging.Infof("[RunAI] User provided CA Chain Path: %s, Custom CA enabled", runaiData.GetInputValue(configinputs.InstallInputCAChainPath))
	}

	if runaiData.GetInputValue(configinputs.InstallInputKubernetesDistribution) != "" {
		k8sDist := runaiData.GetInputValue(configinputs.InstallInputKubernetesDistribution)
		values = append(values, "global.config.kubernetesDistribution="+k8sDist)
		logging.Infof("[RunAI] Setting Kubernetes distribution: %s", k8sDist)
	}

	return values
}

// GetTimeoutHint returns troubleshooting guidance when the step times out.
func (i *InstallCPChart) GetTimeoutHint() string {
	return "Control Plane did not become accessible within timeout - this may indicate deployment issues. Check pod status with: kubectl get pods -n runai-backend"
}
