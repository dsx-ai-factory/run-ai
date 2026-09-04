/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"context"
	"encoding/json"
	"fmt"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SuspendRunner enforces a time/idle limit on a WorkloadRunner by flipping the
// native suspend field of the workload manifest embedded in the runner's
// Spec.Template, then writing the runner back. The dromos controller observes
// the spec change and re-applies the (now-suspended) template to the
// materialized child via its normal server-side-apply, reusing the upstream
// operator's existing suspend semantics. This keeps the runtime-rules blast radius
// to the WorkloadRunner alone — it never reads or mutates the child directly,
// so no RBAC beyond the existing workloadrunners get/patch is required.
//
// The native suspend field differs by wrapped kind:
//   - batch/v1 Job              → spec.suspend
//   - kubeflow.org operator kinds → spec.runPolicy.suspend
//
// The same write clears the runner's start-time annotation so a subsequent
// user-initiated resume is treated as a fresh run with a fresh allowance (the
// lifecycle annotation helper re-stamps start-time when the runner re-enters a
// running phase).
//
// Error semantics (per AC):
//   - a missing runner is treated as already-suspended (success);
//   - an unsupported wrapped kind surfaces a clear error;
//   - the patch uses JSON merge semantics (no resourceVersion precondition), so
//     it is conflict-free.
func SuspendRunner(ctx context.Context, c client.Client, namespace, runnerName string) error {
	var runner dromosv1alpha1.WorkloadRunner
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: runnerName}, &runner); err != nil {
		if apierrors.IsNotFound(err) {
			// Runner gone → nothing to suspend. The wrapped child is cleaned up
			// via owner-reference cascade, so this is a success.
			return nil
		}
		return err
	}

	template, err := parseTemplate(runner.Spec.Template.Raw)
	if err != nil {
		return fmt.Errorf("runner %s/%s: parse spec.template: %w", namespace, runnerName, err)
	}

	suspendPath, err := suspendFieldPath(template.GroupVersionKind())
	if err != nil {
		return fmt.Errorf("runner %s/%s: %w", namespace, runnerName, err)
	}

	alreadySuspended, _, _ := unstructured.NestedBool(template.Object, suspendPath...)
	_, startTimePresent := runner.GetAnnotations()[commonconstants.StartTimeAnnotation]
	if alreadySuspended && !startTimePresent {
		// Idempotent: nothing to change.
		return nil
	}

	// Patch only the suspend field, not the whole template — sending the whole
	// template back would clobber concurrent edits to its other fields.
	templatePatch := map[string]any{}
	if err := unstructured.SetNestedField(templatePatch, true, suspendPath...); err != nil {
		return fmt.Errorf("runner %s/%s: build suspend patch %v: %w", namespace, runnerName, suspendPath, err)
	}

	patchBody, err := json.Marshal(map[string]any{
		"spec": map[string]any{
			"template": templatePatch,
		},
		"metadata": map[string]any{
			"annotations": map[string]any{
				commonconstants.StartTimeAnnotation: nil,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("runner %s/%s: marshal suspend patch: %w", namespace, runnerName, err)
	}

	if err := c.Patch(ctx, &runner, client.RawPatch(types.MergePatchType, patchBody)); err != nil {
		return fmt.Errorf("runner %s/%s: patch suspend: %w", namespace, runnerName, err)
	}
	return nil
}

// suspendFieldPath returns the path to the native suspend field for the wrapped
// child kind identified by gvk. Generic batch Jobs expose spec.suspend; the
// kubeflow distributed-training operator kinds expose spec.runPolicy.suspend.
func suspendFieldPath(gvk schema.GroupVersionKind) ([]string, error) {
	if gvk.Group == "batch" && gvk.Kind == "Job" {
		return []string{"spec", "suspend"}, nil
	}
	if gvk.Group == commonconstants.GroupKubeflow {
		switch gvk.Kind {
		case commonconstants.KindMpiJob,
			commonconstants.KindPyTorchJob,
			commonconstants.KindTFJob,
			commonconstants.KindXGBoostJob,
			commonconstants.KindJAXJob:
			return []string{"spec", "runPolicy", "suspend"}, nil
		}
	}
	return nil, fmt.Errorf("unsupported wrapped kind %s for suspend", gvk)
}
