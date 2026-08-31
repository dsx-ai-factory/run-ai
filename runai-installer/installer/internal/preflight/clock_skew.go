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

func init() {
	builders = append(builders, clockSkewBuilder, clockDriftBuilder)
}

func clockSkewBuilder(_ *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	return steps.NewActionStepTitled("preflight-clock-skew", "Node clock skew", actions.CheckClockSkew{})
}

func clockDriftBuilder(_ *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	return steps.NewActionStepTitled("preflight-clock-drift", "Node clock drift", actions.CheckClockDrift{})
}
