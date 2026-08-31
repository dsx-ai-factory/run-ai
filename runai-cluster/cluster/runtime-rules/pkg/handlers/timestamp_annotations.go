/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileTimestampAnnotations stamps and clears the lifecycle annotations
// (run.ai/start-time, run.ai/completion-time) the runtime-rules owns on a
// WorkloadRunner. Called before the conversion so the converter sees the
// freshly-stamped value on the same event.
//
// State machine:
//
//	Running / Degraded   → set start-time if absent; clear completion-time if present
//	Completed / Failed   → set completion-time if absent; clear start-time if present
//	Suspended / Stopped  → clear both timestamps (reset for a fresh run)
//	any other phase      → no-op
//
// Idempotent: repeated observations of the same phase produce no patch when
// the desired keys are already in their target state. A failed patch logs
// and is retried on the next event — see design §17 risk-5.
//
// Implementation: JSON merge patch (RFC 7396) — adding a key when annotations
// is nil works (merge patch creates intermediate objects), and `null` clears.
// JSON-patch would have failed the "add" with a missing parent path.
func (workloadHandler *WorkloadHandler) ReconcileTimestampAnnotations(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) error {
	annotationsPatch := desiredAnnotationDelta(runner)
	if len(annotationsPatch) == 0 {
		return nil
	}

	body, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"annotations": annotationsPatch,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal merge patch: %w", err)
	}
	patch := client.RawPatch(types.MergePatchType, body)
	if err = workloadHandler.K8sClient.Patch(ctx, runner, patch); err != nil {
		zap.S().Warnf("failed to patch lifecycle annotations on runner %s/%s: %v", runner.Namespace, runner.Name, err)
		return err
	}
	return nil
}

// desiredAnnotationDelta returns the minimal annotation delta to bring the
// runner to the state the phase implies. A value of nil for a key means
// "remove" (JSON merge patch semantics). An empty map means no change.
//
// Visible for testing.
func desiredAnnotationDelta(runner *dromosv1alpha1.WorkloadRunner) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	annotations := runner.GetAnnotations()

	var toSet, toClear []string
	switch runner.Status.Phase {
	case dromosv1alpha1.WorkloadRunnerPhaseRunning, dromosv1alpha1.WorkloadRunnerPhaseDegraded:
		toSet = []string{commonconstants.StartTimeAnnotation}
		toClear = []string{commonconstants.CompletionTimeAnnotation}
	case dromosv1alpha1.WorkloadRunnerPhaseCompleted, dromosv1alpha1.WorkloadRunnerPhaseFailed:
		toSet = []string{commonconstants.CompletionTimeAnnotation}
		toClear = []string{commonconstants.StartTimeAnnotation}
	case dromosv1alpha1.WorkloadRunnerPhaseSuspended, dromosv1alpha1.WorkloadRunnerPhaseStopped:
		toClear = []string{commonconstants.StartTimeAnnotation, commonconstants.CompletionTimeAnnotation}
	default:
		return nil
	}

	delta := map[string]any{}
	for _, key := range toSet {
		if _, present := annotations[key]; !present {
			delta[key] = now
		}
	}
	for _, key := range toClear {
		if _, present := annotations[key]; present {
			delta[key] = nil
		}
	}
	if len(delta) == 0 {
		return nil
	}
	return delta
}
