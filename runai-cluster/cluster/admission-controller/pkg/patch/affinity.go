/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
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
