/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

const PrometheusStepName = "Installing Prometheus"

type PrometheusInstallStep struct {
	*step.RunAIInstallStepImpl
	k8sClient k8s.K8sClients
	version   string
}

func NewPrometheusInstallStep(k8sClient k8s.K8sClients, version string) step.InstallStep {
	p := &PrometheusInstallStep{
		k8sClient:            k8sClient,
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(PrometheusStepName),
		version:              version,
	}
	p.SetStatusEvaluator(p.EvaluateStatus)
	return p
}

func (p *PrometheusInstallStep) EvaluateStatus(runaiData *model.RunAIInstallData) {
	// Early return if already in terminal state (prevents race with timeout detection)
	if p.ShouldSkipStatusEvaluation() {
		return
	}

	logging.Infof("[Prereq] Checking if Prometheus is installed...")
	err := ValidateComponentNamespace(context.Background(), p.k8sClient.GetClient(), "prometheus", "operator", "")
	if err != nil {
		elapsed := p.GetElapsedTime()
		timeout := p.GetTimeout()
		if elapsed > 0 {
			logging.Infof("[Prereq] Prometheus not yet ready (elapsed: %v, timeout: %v)", elapsed, timeout)
		} else {
			logging.Infof("[Prereq] Prometheus not found in cluster")
		}
		return
	}
	logging.Infof("[Prereq] Prometheus is installed")
	p.SetStatus(model.Installed, nil)
}

func (p *PrometheusInstallStep) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	logging.Infof("[Prereq] Installing Prometheus (kube-prometheus-stack) v%s...", p.version)
	p.SetStatus(model.Installing, nil)
	err := helmclient.InstallOrUpgradeRelease(context.Background(),
		"oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack",
		p.version, "kube-prometheus-stack", "kube-prometheus-stack", "monitoring", true, []string{"grafana.enabled=false", "prometheus.prometheusSpec.scrapeConfigs[0].jobLabel=runai-cluster"})
	if err != nil {
		logging.Errorf("[Prereq] Failed to install Prometheus: %v", err)
		p.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Prereq] Prometheus Helm chart deployed, waiting for readiness")
	return nil
}
