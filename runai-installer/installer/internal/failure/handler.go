/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package failure provides a terminal-driven FailureHandler for the
// orchestrator. In interactive mode it prompts the user to choose between
// abort, skip, and retry; in non-interactive mode it follows the configured
// FailurePolicy.
package failure

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
)

const (
	choiceAbort = "abort"
	choiceSkip  = "skip"
	choiceRetry = "retry"
)

type Handler struct {
	term   *terminal.Terminal
	policy spec.FailurePolicy
}

func New(t *terminal.Terminal, policy spec.FailurePolicy) *Handler {
	return &Handler{term: t, policy: policy}
}

func NewFromSpec(t *terminal.Terminal, s *spec.Spec) *Handler {
	return New(t, s.OnStepFailure)
}

func (h *Handler) OnFailure(ctx context.Context, info orchestrator2.FailureInfo) orchestrator2.FailureDecision {
	h.term.Error("step %q in flow %q failed during %s: %v",
		info.Step, info.Flow, info.Phase, info.Err)

	if h.term.Yes() {
		return h.nonInteractive()
	}
	return h.prompt()
}

func (h *Handler) nonInteractive() orchestrator2.FailureDecision {
	if h.policy == spec.FailurePolicySkip {
		h.term.Warning("non-interactive: skipping step (policy=skip)")
		return orchestrator2.FailureSkip
	}
	h.term.Warning("non-interactive: aborting (policy=abort)")
	return orchestrator2.FailureAbort
}

func (h *Handler) prompt() orchestrator2.FailureDecision {
	choice, err := h.term.Select(
		"How do you want to proceed?",
		[]string{choiceAbort, choiceSkip, choiceRetry},
		choiceAbort,
	)
	if err != nil {
		h.term.Warning("prompt failed (%v); aborting", err)
		return orchestrator2.FailureAbort
	}
	switch choice {
	case choiceSkip:
		return orchestrator2.FailureSkip
	case choiceRetry:
		return orchestrator2.FailureRetry
	default:
		return orchestrator2.FailureAbort
	}
}
