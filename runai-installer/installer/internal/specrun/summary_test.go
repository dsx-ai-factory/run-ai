/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package specrun

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func preflightStep(results ...orchestrator2.ActionResult) orchestrator2.Event {
	return orchestrator2.Event{Pipeline: "preflight", Kind: orchestrator2.EventKindStepSucceeded, Results: results}
}

func TestPreflightSummary_Passed(t *testing.T) {
	sum := preflightSummary([]orchestrator2.Event{
		preflightStep(),
		preflightStep(orchestrator2.ActionResult{Data: map[string]any{"kubernetesVersion": "v1.30.0"}}),
	})
	assert.Equal(t, "passed", sum.Outcome)
	assert.Equal(t, 2, sum.Passed)
	assert.Zero(t, sum.Warned)
	assert.Zero(t, sum.Failed)
}

func TestPreflightSummary_Warning(t *testing.T) {
	sum := preflightSummary([]orchestrator2.Event{
		preflightStep(),
		preflightStep(orchestrator2.ActionResult{Status: orchestrator2.StatusWarn, Reason: "deprecated setting"}),
	})
	assert.Equal(t, "warning", sum.Outcome)
	assert.Equal(t, 1, sum.Passed)
	assert.Equal(t, 1, sum.Warned)
	assert.Zero(t, sum.Failed)
}

func TestPreflightSummary_FailBeatsWarnWithinAStep(t *testing.T) {
	sum := preflightSummary([]orchestrator2.Event{
		preflightStep(orchestrator2.ActionResult{Status: orchestrator2.StatusWarn}),
		preflightStep(
			orchestrator2.ActionResult{Status: orchestrator2.StatusWarn},
			orchestrator2.ActionResult{Status: orchestrator2.StatusFail},
		),
	})
	assert.Equal(t, "failed", sum.Outcome)
	assert.Equal(t, 1, sum.Failed)
	assert.Equal(t, 1, sum.Warned)
}

func TestPreflightSummary_StepFailedCountsAsFailure(t *testing.T) {
	// A hard-errored preflight check (step_failed, no verdict results) must not
	// be silently dropped and reported as passed.
	sum := preflightSummary([]orchestrator2.Event{
		preflightStep(),
		{Pipeline: "preflight", Kind: orchestrator2.EventKindStepFailed},
	})
	assert.Equal(t, "failed", sum.Outcome)
	assert.Equal(t, 1, sum.Passed)
	assert.Equal(t, 1, sum.Failed)
}

func TestPreflightSummary_IgnoresOtherPipelines(t *testing.T) {
	sum := preflightSummary([]orchestrator2.Event{
		{Pipeline: "runai", Kind: orchestrator2.EventKindStepSucceeded, Results: []orchestrator2.ActionResult{{Status: orchestrator2.StatusFail}}},
		{Pipeline: "runai", Kind: orchestrator2.EventKindStepFailed},
		preflightStep(),
	})
	assert.Equal(t, "passed", sum.Outcome)
	assert.Equal(t, 1, sum.Passed)
	assert.Zero(t, sum.Failed)
}

func TestControlPlaneSummary_AdminCredentials(t *testing.T) {
	s := &spec.Spec{ControlPlane: &spec.ControlPlane{Domain: "runai.example.com"}}

	cp := controlPlaneSummary(s, "2.26.0", true)
	require.NotNil(t, cp.AdminCredentials)
	assert.Equal(t, "runai-backend", cp.AdminCredentials.Namespace)
	assert.Equal(t, "runai-backend-tenants-manager", cp.AdminCredentials.UsernameConfigMap)
	assert.Equal(t, "ADMIN_USERNAME", cp.AdminCredentials.UsernameKey)
	assert.Equal(t, "runai-backend-gen-tenants-manager", cp.AdminCredentials.PasswordSecret)
	assert.Equal(t, "ADMIN_PASSWORD", cp.AdminCredentials.PasswordKey)

	assert.Equal(t, "runai-backend-tenants-manager", controlPlaneSummary(s, "2.25.0", true).AdminCredentials.PasswordSecret)
	assert.Nil(t, controlPlaneSummary(s, "2.26.0", false).AdminCredentials)
}
