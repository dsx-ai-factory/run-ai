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

// This file reads the readiness signal already computed by
// github.com/kai-scheduler/kai-gpu-fractioning (the GPU Sharing Operator),
// instead of independently re-deriving GPU Operator version support. That
// operator's own GpuOperatorDependencyChecker already reads the NVIDIA GPU
// Operator's ClusterPolicy/ClusterServiceVersion, validates its version
// against the same minimum CUDA-fractions requires, and publishes the result
// as the "Ready" condition on its singleton GpuFractioningConfig CR (named
// "default", cluster-scoped). We read that condition rather than duplicating
// the version-detection and semver logic ourselves.
//
// kai-gpu-fractioning is not currently vendored as a Go dependency of this
// repo (see internal/run/consts and go.mod), so this reads the CR via the
// unstructured client and a well-known GVK, the same pattern already used
// for GetDependencyStatus/ClusterPolicy - no new go.mod dependency, and it
// degrades gracefully (Available: true, i.e. non-blocking) if the CRD isn't
// installed on the cluster.

const gpuFractioningConfigName = "default"

var gpuFractioningConfigGVK = schema.GroupVersionKind{
	Group:   "gpu-fractioning.kai.scheduler",
	Version: "v1alpha1",
	Kind:    "GpuFractioningConfig",
}

// CheckGpuFractioningReadiness reads the singleton GpuFractioningConfig CR's
// aggregate "Ready" condition, published by the GPU Sharing Operator
// (kai-gpu-fractioning). It reflects GPU Operator version support as well as
// GPU driver and daemon health for CUDA fractions.
//
// If the CRD is not installed, or the "default" CR does not exist yet, this
// is treated as "unknown" and reported as available (non-blocking) - the
// GPU Sharing Operator not being deployed is not itself evidence that GPU
// fractions prerequisites are unmet.
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
		// No Ready condition yet (e.g. CR just created, controller has not
		// reconciled it). Not evidence of a missing prerequisite.
		return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
	}

	if status != "False" {
		return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
	}

	if reason == "" {
		// A False Ready condition always means fractions are not usable, even
		// if the message is empty - do not surface that as a silent, reasonless
		// MissingPrerequisites.
		reason = "GPU fractions are not supported on this cluster: kai-gpu-fractioning reports Ready=False with no message"
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(false),
		Reason:    reason,
	}
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
		condStatus, _, _ := unstructured.NestedString(condMap, "status")
		condMessage, _, _ := unstructured.NestedString(condMap, "message")
		return condStatus, condMessage, true
	}

	return "", "", false
}
