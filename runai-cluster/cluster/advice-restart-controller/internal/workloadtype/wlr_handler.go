/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"encoding/json"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var GVKWorkloadRunner = schema.GroupVersionKind{
	Group:   dromosv1alpha1.GroupVersion.Group,
	Version: dromosv1alpha1.GroupVersion.Version,
	Kind:    dromosv1alpha1.WorkloadRunnerKind,
}

// nodeAffinityOverridePath is the only spec override path the RestartRequest validator admits.
const nodeAffinityOverridePath = "/spec/instructions/nodeAffinity"

// WorkloadRunnerHandler handles WorkloadRunner top owners.
type WorkloadRunnerHandler struct {
	wlr *dromosv1alpha1.WorkloadRunner
}

// NewWorkloadRunnerHandler converts an unstructured object to a typed WorkloadRunner and returns a handler.
func NewWorkloadRunnerHandler(obj *unstructured.Unstructured) (WorkloadHandler, error) {
	wlr := &dromosv1alpha1.WorkloadRunner{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, wlr); err != nil {
		return nil, fmt.Errorf("convert to WorkloadRunner: %w", err)
	}
	return &WorkloadRunnerHandler{wlr: wlr}, nil
}

func (h *WorkloadRunnerHandler) GetPhase() string {
	return string(h.wlr.Status.Phase)
}

// BuildSpecOverrides emits a single add of the whole node-affinity instruction, carrying the
// union of the nodes already excluded there and the new exclusion
func (h *WorkloadRunnerHandler) BuildSpecOverrides(exclusions []string) (runv1alpha1.SpecOverrides, error) {
	if len(exclusions) == 0 {
		return nil, nil
	}

	excluded := append(h.existingExcludedNodes(), exclusions...)
	slices.Sort(excluded)
	excluded = slices.Compact(excluded)

	affinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{{
				MatchExpressions: []corev1.NodeSelectorRequirement{{
					Key:      corev1.LabelHostname,
					Operator: corev1.NodeSelectorOpNotIn,
					Values:   excluded,
				}},
			}},
		},
	}

	data, err := json.Marshal(dromosv1alpha1.NodeAffinityConfig{NodeAffinity: affinity})
	if err != nil {
		return nil, err
	}

	return runv1alpha1.SpecOverrides{{
		Op:    runv1alpha1.JSONPatchOperationAdd,
		Path:  nodeAffinityOverridePath,
		Value: &runtime.RawExtension{Raw: data},
	}}, nil
}

// existingExcludedNodes returns the hostnames the workload already excludes
func (h *WorkloadRunnerHandler) existingExcludedNodes() []string {
	instructions := h.wlr.Spec.Instructions
	if instructions == nil || instructions.NodeAffinity == nil || instructions.NodeAffinity.NodeAffinity == nil {
		return nil
	}

	required := instructions.NodeAffinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil {
		return nil
	}

	var nodes []string
	for _, term := range required.NodeSelectorTerms {
		for _, expr := range term.MatchExpressions {
			if expr.Key == corev1.LabelHostname && expr.Operator == corev1.NodeSelectorOpNotIn {
				nodes = append(nodes, expr.Values...)
			}
		}
	}
	return nodes
}
