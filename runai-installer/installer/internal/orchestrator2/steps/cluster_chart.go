/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// clusterChartAction installs the cluster helm release and exposes the control
// plane coordinates PostValidate needs, whether the chart is remote
// (actions.InstallClusterChart) or local (actions.InstallClusterChartLocal).
type clusterChartAction interface {
	chartReleaseAction
	ClusterRef() (clusterName string, controlPlane actions.ControlPlaneRef)
	UsesSuppliedCredentials() bool
}

// clusterChartStep installs the Run:ai cluster Helm chart. Detect reports the
// chart as installed when its release exists; PostValidate polls the control
// plane until the cluster reports a recent liveness.
type clusterChartStep struct {
	*orchestrator2.BaseStep
	deps   orchestrator2.Deps
	action clusterChartAction
}

func NewClusterChart(name string, action clusterChartAction, deps orchestrator2.Deps) orchestrator2.Step {
	return &clusterChartStep{BaseStep: orchestrator2.NewBaseStep(name), deps: deps, action: action}
}

func (s *clusterChartStep) Detect(_ context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	release, namespace := s.action.Release()
	return helmReleaseDetect(s.deps, release, namespace)
}

func (s *clusterChartStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{s.action}, nil
}

func (s *clusterChartStep) PostValidate(ctx context.Context, _ *spec.Spec) error {
	if s.action.UsesSuppliedCredentials() {
		return nil
	}
	clusterName, controlPlane := s.action.ClusterRef()
	return pollUntil(ctx, func(ctx context.Context) bool {
		apiClient, err := controlPlane.Client(s.deps)
		if err != nil {
			return false
		}
		uuid, err := actions.FindClusterUUID(ctx, apiClient, clusterName)
		if err != nil || uuid == "" {
			return false
		}
		dc, status, err := apiClient.GetClusterByUUID(ctx, uuid)
		if err != nil || status != 200 || !dc.HasLastLiveness() {
			return false
		}
		ll := dc.GetLastLiveness()
		return ll.After(time.Now().Add(-time.Minute)) && ll.Before(time.Now())
	})
}
