/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() { builders = append(builders, openShiftMonitoringBuilder) }

// openShiftMonitoringBuilder assembles the OpenShift monitoring check. It only
// applies on OpenShift (where Run:ai uses the built-in monitoring stack instead
// of the Prometheus prerequisite) and needs a ClusterProbe, so it returns nil
// otherwise.
func openShiftMonitoringBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if !s.IsOpenShift() || deps.ClusterProbe == nil {
		return nil
	}
	return steps.NewActionStepTitled("preflight-openshift-monitoring", "OpenShift monitoring stack", actions.CheckOpenShiftMonitoring{})
}
