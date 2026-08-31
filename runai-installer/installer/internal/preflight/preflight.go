/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package preflight assembles the preflight flow: read-only checks that run
// against the live cluster (and, for air-gap, the registry) before any install
// work. Each check is implemented as a read-only action (in the actions
// package) wrapped in a step; this package only decides which checks apply and
// orders them into a flow.
//
// The preflight flow runs in its own pipeline that executes in every mode,
// including --dry-run, because the checks make no changes. Checks report their
// verdict via the action's ActionResult (pass/warn/fail) rather than aborting,
// so every check runs and the results are aggregated by the caller.
package preflight

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

const flowName = "preflight-checks"

// builder produces the step for one check, or nil when the check does not apply
// to the given spec (e.g. registry checks only for air-gapped installs).
type builder func(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step

// builders is the ordered registry of checks. Each check file appends its
// builder in an init function.
var builders []builder

// BuildFlow assembles the preflight flow for the spec, one step per applicable
// check. It returns nil when no checks apply, so the caller can omit an empty
// preflight flow.
func BuildFlow(s *spec.Spec, deps orchestrator2.Deps) *orchestrator2.Flow {
	var steps []orchestrator2.Step
	for _, b := range builders {
		if step := b(s, deps); step != nil {
			steps = append(steps, step)
		}
	}
	if len(steps) == 0 {
		return nil
	}
	return &orchestrator2.Flow{Name: flowName, Title: "Running checks", Steps: steps}
}
