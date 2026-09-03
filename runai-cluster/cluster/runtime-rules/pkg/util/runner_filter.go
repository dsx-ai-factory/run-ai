/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"go.uber.org/zap"
)

// RunnerSkipReason names a single, observable cause for dropping a
// WorkloadRunner from the enforcement pipeline. The string form lands in the
// debug-level log line emitted by the handler, so keep it stable for grepping.
type RunnerSkipReason string

const (
	RunnerSkipNotV1Cast           RunnerSkipReason = "not-v1-cast"
	RunnerSkipMissingWorkloadKind RunnerSkipReason = "missing-workload-kind"
	RunnerSkipOutOfScopeKind      RunnerSkipReason = "out-of-scope-kind"
	RunnerSkipTerminalPhase       RunnerSkipReason = "terminal-phase"
	RunnerSkipPreRunningPhase     RunnerSkipReason = "pre-running-phase"
)

// ClassifyRunner runs the four scope checks the design doc §4 demands, in
// the documented order. On success it returns the legacy workload kind the
// runner should use for project limits. On failure it returns the reason for skipping
// so the caller can log a structured drop reason.
//
// Order matters: cast-label first (cheapest gate, biggest culling factor),
// then wrapped-kind label (drops the out-of-scope kinds before we even look
// at status), then phase (status-dependent so it may flap; checked last so
// successive phase transitions all emit a consistent skip stream).
func ClassifyRunner(runner *dromosv1alpha1.WorkloadRunner) (categoryKind string, skip RunnerSkipReason, ok bool) {
	templateLabels := runnerTemplateLabels(runner)
	if templateLabels[constants.ConvertedFromNativeLabel] != "true" {
		return "", RunnerSkipNotV1Cast, false
	}

	wrappedKind, hasWrappedKind := templateLabels[commonconstants.LabelWorkloadKind]
	if !hasWrappedKind || wrappedKind == "" {
		return "", RunnerSkipMissingWorkloadKind, false
	}

	category, inScope := categoryForWrappedKind(wrappedKind)
	if !inScope {
		return "", RunnerSkipOutOfScopeKind, false
	}

	switch runner.Status.Phase {
	case dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseDegraded:
		return category, "", true
	case dromosv1alpha1.WorkloadRunnerPhaseCreating,
		dromosv1alpha1.WorkloadRunnerPhasePending,
		dromosv1alpha1.WorkloadRunnerPhaseInitializing:
		return "", RunnerSkipPreRunningPhase, false
	default:
		// Terminal-style: Completed, Failed, Stopped, Suspended, Unknown, plus
		// any future phase added to the CRD. Fail-closed mirrors the existing
		// isRunnerActivePhase helper.
		return "", RunnerSkipTerminalPhase, false
	}
}

// runnerTemplateLabels returns the metadata.labels of the manifest embedded in
// the runner's Spec.Template, where the control-plane stamps the cast marker and
// workloadKind. Returns nil (so every lookup misses — fail-closed) when the
// template is empty or unparseable.
func runnerTemplateLabels(runner *dromosv1alpha1.WorkloadRunner) map[string]string {
	tmpl, err := parseTemplate(runner.Spec.Template.Raw)
	if err != nil {
		return nil
	}
	return tmpl.GetLabels()
}

// IsPhaseSkip reports whether a skip reason is purely phase-based — the runner
// is in scope (ours: correct cast label and an in-scope wrapped kind) but is
// not currently in an enforceable (Running/Degraded) phase. Scope-based skips
// (not-v1-cast, missing/out-of-scope wrapped kind) mean the runner isn't ours
// at all, so the runtime-rules must leave its annotations untouched.
func IsPhaseSkip(reason RunnerSkipReason) bool {
	return reason == RunnerSkipTerminalPhase || reason == RunnerSkipPreRunningPhase
}

// categoryForWrappedKind maps the `workloadKind` label values the control-plane
// writes during the v1->v2 cast to the project-limit bucket that governs the runner.
//
// The Workspace -> Interactive and Training -> Training mappings cover the
// generic Job-wrapped case. The five distributed-training operator kinds all
// use the distributed-training bucket, which shares the training limits.
func categoryForWrappedKind(wrappedKind string) (string, bool) {
	switch wrappedKind {
	case "Workspace":
		return commonconstants.KindInteractiveWorkload, true
	case "Training":
		return commonconstants.KindTrainingWorkload, true
	case commonconstants.KindMpiJob,
		commonconstants.KindPyTorchJob,
		commonconstants.KindTFJob,
		commonconstants.KindXGBoostJob,
		commonconstants.KindJAXJob:
		return commonconstants.KindDistributedWorkload, true
	default:
		return "", false
	}
}

// WarnIfMissingWorkloadID emits a warning when the runner lacks the
// control-plane's workload-id annotation. Per AC bullet 5, this is a
// warning only — the runner is NOT skipped. The annotation is non-critical for
// enforcement (the runner UID is still a valid fallback inside GetWorkloadID),
// but its absence indicates a control-plane state drift worth surfacing.
func WarnIfMissingWorkloadID(runner *dromosv1alpha1.WorkloadRunner) {
	if _, ok := runner.Annotations[constants.ControlPlaneWorkloadIdAnnotationForWorkload]; !ok {
		zap.S().Warnf("WorkloadRunner %s/%s is missing the %q annotation; enforcement will use the runner UID as the workload id",
			runner.Namespace, runner.Name, constants.ControlPlaneWorkloadIdAnnotationForWorkload)
	}
}
