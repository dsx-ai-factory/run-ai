/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

// chartReleaseAction is a perform action that installs a named helm release,
// whether from a remote repo (actions.HelmInstall) or a local bundle chart
// (actions.HelmInstallLocal). It lets a chart step detect prior installs
// independently of where the chart comes from.
type chartReleaseAction interface {
	orchestrator2.PerformAction
	Release() (name, namespace string)
}

// cpChartStep installs the Run:ai control-plane Helm chart. Detect reports the
// chart as installed when its release exists; PostValidate polls the control
// plane API until it answers.
type cpChartStep struct {
	*orchestrator2.BaseStep
	deps   orchestrator2.Deps
	action chartReleaseAction
	caPath string
}

func NewControlPlaneChart(name string, action chartReleaseAction, caPath string, deps orchestrator2.Deps) orchestrator2.Step {
	return &cpChartStep{BaseStep: orchestrator2.NewBaseStep(name), deps: deps, action: action, caPath: caPath}
}

func (s *cpChartStep) Detect(_ context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	release, namespace := s.action.Release()
	return helmReleaseDetect(s.deps, release, namespace)
}

func (s *cpChartStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{s.action}, nil
}

func (s *cpChartStep) PostValidate(ctx context.Context, sp *spec.Spec) error {
	domain := sp.ControlPlane.Domain
	return pollUntil(ctx, func(ctx context.Context) bool {
		apiClient, err := s.deps.RunAIClientProvider(actions.ControlPlaneURL(domain), s.caPath)
		if err != nil {
			return false
		}
		_, status, err := apiClient.GetClusters(ctx)
		return err == nil && status == 200
	})
}
