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
		// Ready is also True when the fractioning daemons have no node to run on (their node selector
		// matches none), and then no node can run fractions: report that instead of "available", or
		// every GPU node shows a missing fractions prerequisite and counts as not ready.
		noTargetNodes, err := daemonsHaveNoTargetNodes(config)
		if err != nil {
			return runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    fmt.Sprintf("unable to read GpuFractioningConfig %q conditions: %v", gpuFractioningConfigName, err),
			}
		}
		if noTargetNodes {
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

// daemonsHaveNoTargetNodes reports whether both fractioning daemons have no node to run on.
func daemonsHaveNoTargetNodes(config *unstructured.Unstructured) (bool, error) {
	fractiondReason, err := conditionReason(config, "FractiondReady")
	if err != nil {
		return false, err
	}
	mpsdReason, err := conditionReason(config, "MpsdReady")
	if err != nil {
		return false, err
	}
	return fractiondReason == noTargetNodesReason && mpsdReason == noTargetNodesReason, nil
}

// conditionReason returns the reason of the condition of that type, or "" when there is none; a type or
// reason that is not a string is an error.
func conditionReason(config *unstructured.Unstructured, conditionType string) (string, error) {
	conditions, ok, err := unstructured.NestedSlice(config.Object, "status", "conditions")
	if err != nil || !ok {
		return "", err
	}
	for _, c := range conditions {
		condMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, err := unstructured.NestedString(condMap, "type")
		if err != nil {
			return "", fmt.Errorf("condition type: %w", err)
		}
		if condType != conditionType {
			continue
		}
		reason, _, err := unstructured.NestedString(condMap, "reason")
		if err != nil {
			return "", fmt.Errorf("condition %s reason: %w", conditionType, err)
		}
		return reason, nil
	}
	return "", nil
}

func readyCondition(config *unstructured.Unstructured) (status, message string, found bool) {
	conditions, ok, err := unstructured.NestedSlice(config.Object, "status", "conditions")
	if err != nil || !ok {
		return "", "", false
	}

	for _, c := range conditions {
		condMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(condMap, "type")
		if condType != "Ready" {
			continue
		}
		condStatus, statusFound, err := unstructured.NestedString(condMap, "status")
		if err != nil || !statusFound || !validConditionStatuses[condStatus] {
			return "", "", false
		}
		condMessage, _, _ := unstructured.NestedString(condMap, "message")
		return condStatus, condMessage, true
	}

	return "", "", false
}
