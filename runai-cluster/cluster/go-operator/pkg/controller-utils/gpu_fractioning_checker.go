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
