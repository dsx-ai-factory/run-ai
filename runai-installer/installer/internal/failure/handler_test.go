/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package failure

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/terminal"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func init() {
	pterm.DisableColor()
	pterm.DisableStyling()
}

func newTermWithYes() (*terminal.Terminal, *bytes.Buffer, *bytes.Buffer) {
	var out, errBuf bytes.Buffer
	t := terminal.New(
		terminal.WithWriter(&out),
		terminal.WithErrWriter(&errBuf),
		terminal.WithYes(true),
	)
	return t, &out, &errBuf
}

func sampleInfo() orchestrator2.FailureInfo {
	return orchestrator2.FailureInfo{
		Pipeline: "install",
		Flow:     "prerequisites",
		Step:     "prometheus",
		Phase:    "perform",
		Err:      errors.New("boom"),
	}
}

func TestHandler_NonInteractive_DefaultAborts(t *testing.T) {
	term, _, errBuf := newTermWithYes()
	h := New(term, spec.FailurePolicyAbort)

	got := h.OnFailure(context.Background(), sampleInfo())

	assert.Equal(t, orchestrator2.FailureAbort, got)
	assert.Contains(t, errBuf.String(), `step "prometheus" in flow "prerequisites" failed during perform: boom`)
}

func TestHandler_NonInteractive_PolicySkip(t *testing.T) {
	term, out, _ := newTermWithYes()
	h := New(term, spec.FailurePolicySkip)

	got := h.OnFailure(context.Background(), sampleInfo())

	assert.Equal(t, orchestrator2.FailureSkip, got)
	assert.Contains(t, out.String(), "skipping step (policy=skip)")
}

func TestHandler_NewFromSpec_UsesSpecPolicy(t *testing.T) {
	term, _, _ := newTermWithYes()
	s := &spec.Spec{NonInteractive: true, OnStepFailure: spec.FailurePolicySkip}

	h := NewFromSpec(term, s)
	got := h.OnFailure(context.Background(), sampleInfo())

	assert.Equal(t, orchestrator2.FailureSkip, got)
}
