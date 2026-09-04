/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// ActionStep is a generic step that performs a fixed list of actions and has no
// detection of its own (BaseStep's Detect always reports not-installed). Use it
// for steps with no meaningful "already done" check; prefer a typed step with a
// real Detect when one is available.
type ActionStep struct {
	*orchestrator2.BaseStep
	actions []orchestrator2.PerformAction
}

func NewActionStep(name string, acts ...orchestrator2.PerformAction) *ActionStep {
	return &ActionStep{BaseStep: orchestrator2.NewBaseStep(name), actions: acts}
}

// NewActionStepTitled is NewActionStep with a custom human-readable title for
// the text UI (the machine name stays the identity used in JSON and tests).
func NewActionStepTitled(name, title string, acts ...orchestrator2.PerformAction) *ActionStep {
	return &ActionStep{BaseStep: orchestrator2.NewBaseStepTitled(name, title), actions: acts}
}

// titleSetter is implemented by every step through the embedded BaseStep.
type titleSetter interface{ SetTitle(string) }

// Titled assigns a custom human-readable title to a step and returns it, so the
// plan can label a step built by any typed constructor without threading a
// title parameter through it. The machine name is unchanged.
func Titled(s orchestrator2.Step, title string) orchestrator2.Step {
	if ts, ok := s.(titleSetter); ok {
		ts.SetTitle(title)
	}
	return s
}

// extraTimeoutSetter is implemented by every step through the embedded BaseStep.
type extraTimeoutSetter interface{ SetExtraTimeout(time.Duration) }

// WithExtraTimeout sets a step's extra timeout (added to the configured per-step
// timeout by the runner) and returns it, mirroring Titled.
func WithExtraTimeout(s orchestrator2.Step, d time.Duration) orchestrator2.Step {
	if es, ok := s.(extraTimeoutSetter); ok {
		es.SetExtraTimeout(d)
	}
	return s
}

func (s *ActionStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return s.actions, nil
}
