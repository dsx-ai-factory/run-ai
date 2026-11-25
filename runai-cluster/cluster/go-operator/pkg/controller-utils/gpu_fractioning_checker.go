/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"context"
	"fmt"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const gpuFractioningConfigName = "default"

var gpuFractioningConfigGVK = schema.GroupVersionKind{
	Group:   "gpu-fractioning.kai.scheduler",
	Version: "v1alpha1",
	Kind:    "GpuFractioningConfig",
}

func CheckGpuFractioningReadiness(ctx context.Context, reader client.Reader) runv1.PublicStatus {
	config := &unstructured.Unstructured{}
	config.SetGroupVersionKind(gpuFractioningConfigGVK)

	if err := reader.Get(ctx, types.NamespacedName{Name: gpuFractioningConfigName}, config); err != nil {
		if meta.IsNoMatchError(err) || kerrors.IsNotFound(err) {
			return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
		}
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("unable to read GpuFractioningConfig %q: %v", gpuFractioningConfigName, err),
		}
	}

	status, reason, found := readyCondition(config)
	if !found {
		return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
	}

	if status != "False" {
		// Ready is also True when both fractioning daemons have no node to run on (their node selector
		// matches none, e.g. no node meets the driver requirement). Then no node can run fractions:
		// report that instead of "available", or every GPU node shows a missing fractions prerequisite.
		if daemonsHaveNoTargetNodes(ctx, config) {
			return runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    "GPU fractions are not deployed on any node: kai-gpu-fractioning has no target nodes",
			}
		}
		return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
	}

	if reason == "" {
		reason = "GPU fractions are not supported on this cluster: kai-gpu-fractioning reports Ready=False with no message"
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(false),
		Reason:    reason,
	}
}

var validConditionStatuses = map[string]bool{"True": true, "False": true, "Unknown": true}

const noTargetNodesReason = "NoTargetNodes"

// fractioningDaemonConditions are the per-daemon conditions of the GpuFractioningConfig.
var fractioningDaemonConditions = []string{"FractiondReady", "MpsdReady"}

type condition struct {
	status, reason, message string
	// unreadable is set when status, reason or message is present but not a string.
	unreadable bool
}

// findCondition returns the status condition of the given type. Entries that are not objects, or whose
// type is not a string, are skipped.
func findCondition(config *unstructured.Unstructured, conditionType string) (condition, bool) {
	conditions, ok, err := unstructured.NestedSlice(config.Object, "status", "conditions")
	if err != nil || !ok {
		return condition{}, false
	}
	for _, c := range conditions {
		condMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if condType, _, err := unstructured.NestedString(condMap, "type"); err != nil || condType != conditionType {
			continue
		}
		status, _, statusErr := unstructured.NestedString(condMap, "status")
		reason, _, reasonErr := unstructured.NestedString(condMap, "reason")
		message, _, messageErr := unstructured.NestedString(condMap, "message")
		return condition{
			status:     status,
			reason:     reason,
			message:    message,
			unreadable: statusErr != nil || reasonErr != nil || messageErr != nil,
		}, true
	}
	return condition{}, false
}

// daemonsHaveNoTargetNodes reports whether every fractioning daemon reports NoTargetNodes. An unreadable
// daemon condition counts as not NoTargetNodes (fail open, like the rest of this check) and is logged.
func daemonsHaveNoTargetNodes(ctx context.Context, config *unstructured.Unstructured) bool {
	for _, conditionType := range fractioningDaemonConditions {
		cond, found := findCondition(config, conditionType)
		if cond.unreadable {
			log.FromContext(ctx).Info("ignoring unreadable GpuFractioningConfig condition",
				"name", gpuFractioningConfigName, "condition", conditionType)
			return false
		}
		if !found || cond.reason != noTargetNodesReason {
			return false
		}
	}
	return true
}

// readyCondition returns the Ready condition's status and message; found is false when there is no
// Ready condition or its status is not a valid condition status.
func readyCondition(config *unstructured.Unstructured) (status, message string, found bool) {
	ready, found := findCondition(config, "Ready")
	if !found || !validConditionStatuses[ready.status] {
		return "", "", false
	}
	return ready.status, ready.message, true
}
