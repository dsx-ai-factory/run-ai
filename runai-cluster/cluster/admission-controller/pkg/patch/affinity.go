/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"
)

const (
	DefaultNodePoolName = "default"
	NodePoolNameLabel   = "runai/node-pool"
	NodePoolDeleting    = "Deleting"
)

// AddNodeAffinityToPodSpec matchExpressionsToAdd are added to the current node affinity with an AND operator,
// while between those requirement there is an OR.
func AddNodeAffinityToPodSpec(podSpec *v1.PodSpec, matchExpressionsToAdd []v1.NodeSelectorRequirement) {
	if podSpec.Affinity == nil {
		podSpec.Affinity = &v1.Affinity{}
	}
	if podSpec.Affinity.NodeAffinity == nil {
		podSpec.Affinity.NodeAffinity = &v1.NodeAffinity{}
	}
	if podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
	}
	selector := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if selector.NodeSelectorTerms == nil || len(selector.NodeSelectorTerms) == 0 {
		selector.NodeSelectorTerms = []v1.NodeSelectorTerm{}
		for _, matchExpressionToAdd := range matchExpressionsToAdd {
			selector.NodeSelectorTerms = append(selector.NodeSelectorTerms, v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{matchExpressionToAdd},
			})
		}
		return
	}

	updatedTerms := make([]v1.NodeSelectorTerm, 0)
	for _, matchExpressionToAdd := range matchExpressionsToAdd {
		for _, term := range selector.NodeSelectorTerms {
			updatedExpressions := make([]v1.NodeSelectorRequirement, 0)
			for _, expr := range term.MatchExpressions {
				if expr.Key != matchExpressionToAdd.Key {
					updatedExpressions = append(updatedExpressions, expr)
				}
			}
			updatedExpressions = append(updatedExpressions, matchExpressionToAdd)
			term.MatchExpressions = updatedExpressions
			updatedTerms = append(updatedTerms, term)
		}
	}
	if len(updatedTerms) > 0 {
		selector.NodeSelectorTerms = updatedTerms
	}
}

// HasAcceleratorSelection recognizes both model and resource-only selection.
func HasAcceleratorSelection(spec *v1.PodSpec, annotations map[string]string) bool {
	if annotations[constants.AnnotationGPUTypeLabel] != "" || annotations[constants.AnnotationGPUPlacementResource] != "" {
		return true
	}
	for key := range spec.NodeSelector {
		if strings.HasPrefix(key, constants.AcceleratorResourceLabelPrefix) {
			return true
		}
	}
	// Knative may allow node affinity but disable the Pod nodeSelector field.
	if spec.Affinity == nil || spec.Affinity.NodeAffinity == nil {
		return false
	}
	required := spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil {
		return false
	}
	for _, term := range required.NodeSelectorTerms {
		for _, expression := range term.MatchExpressions {
			if strings.HasPrefix(expression.Key, constants.AcceleratorResourceLabelPrefix) {
				return true
			}
		}
	}
	return false
}

// AddPlacementRequirements applies an OR of placement requirements to a pod. For
// workloads with explicit GPU selection, each existing term must still hold,
// even when a pool or scheduling rule uses the same label as the GPU model.
func AddPlacementRequirements(spec *v1.PodSpec, annotations map[string]string, alternatives []v1.NodeSelectorRequirement) {
	if len(alternatives) == 0 {
		return
	}
	if !HasAcceleratorSelection(spec, annotations) {
		AddNodeAffinityToPodSpec(spec, alternatives)
		return
	}
	if spec.Affinity == nil {
		spec.Affinity = &v1.Affinity{}
	}
	if spec.Affinity.NodeAffinity == nil {
		spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
	}
	affinity := spec.Affinity.NodeAffinity
	if affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		terms := make([]v1.NodeSelectorTerm, 0, len(alternatives))
		for _, req := range alternatives {
			terms = append(terms, v1.NodeSelectorTerm{MatchExpressions: []v1.NodeSelectorRequirement{req}})
		}
		affinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{NodeSelectorTerms: terms}
		return
	}

	required := affinity.RequiredDuringSchedulingIgnoredDuringExecution
	terms := make([]v1.NodeSelectorTerm, 0, len(required.NodeSelectorTerms))
	for _, term := range required.NodeSelectorTerms {
		// Empty terms match no nodes. Adding an expression would open them up.
		if len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0 || termContainsAlternative(term, alternatives) {
			terms = append(terms, term)
			continue
		}
		for _, req := range alternatives {
			combined := term.DeepCopy()
			combined.MatchExpressions = append(combined.MatchExpressions, req)
			terms = append(terms, *combined)
		}
	}
	required.NodeSelectorTerms = terms
}

func termContainsAlternative(term v1.NodeSelectorTerm, alternatives []v1.NodeSelectorRequirement) bool {
	for _, existing := range term.MatchExpressions {
		for _, req := range alternatives {
			if existing.Key == req.Key && existing.Operator == req.Operator && slices.Equal(existing.Values, req.Values) {
				return true
			}
		}
	}
	return false
}

func AddTolerationToSpec(podSpec *v1.PodSpec, key, value string, operator v1.TolerationOperator,
	effect v1.TaintEffect) {
	if len(podSpec.Tolerations) == 0 {
		podSpec.Tolerations = make([]v1.Toleration, 0)
	}
	toleration := v1.Toleration{
		Key:      key,
		Operator: operator,
		Effect:   effect,
		Value:    value,
	}
	for _, existingToleration := range podSpec.Tolerations {
		if existingToleration.MatchToleration(&toleration) {
			return // The toleration is already exists
		}
	}
	podSpec.Tolerations = append(podSpec.Tolerations, toleration)
}
