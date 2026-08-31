/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

// WorkloadHandler provides typed access to a workload's phase and spec overrides.
// Created by a HandlerFactory from an unstructured object.
type WorkloadHandler interface {
	// GetPhase returns the current lifecycle phase of the workload (e.g. "Running", "Failed").
	GetPhase() string
	// BuildSpecOverrides builds SpecOverrides that apply the given node exclusions to the workload.
	BuildSpecOverrides(exclusions []string) (runv1alpha1.SpecOverrides, error)
}

// HandlerFactory creates a WorkloadHandler from an unstructured object.
type HandlerFactory func(obj *unstructured.Unstructured) (WorkloadHandler, error)
