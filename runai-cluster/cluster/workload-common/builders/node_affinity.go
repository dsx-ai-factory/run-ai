/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/utils"
	corev1 "k8s.io/api/core/v1"
)

// Node pool resolution and node-pool-derived affinity live in the sibling
// builders/nodepool package, which the control plane must not compile.
// See that package's doc comment for why.

// MergeNodeAffinity merges the node affinity the user provided with the one we generated internally
func MergeNodeAffinity(userNodeAffinityFields *field_types.NodeAffinityFields, generatedAffinity **corev1.Affinity) {
	if userNodeAffinityFields == nil {
		return
	}

	// if the user provided node affinity, we need to merge it with the one we created
	userAffinity := userNodeAffinityFields.ToNodeAffinity()
	if userAffinity != nil {

		// make sure we don't panic if we didn't generate our affinity
		if *generatedAffinity == nil {
			*generatedAffinity = &corev1.Affinity{}
		}
		if (*generatedAffinity).NodeAffinity == nil {
			(*generatedAffinity).NodeAffinity = &corev1.NodeAffinity{}
		}

		(*generatedAffinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = MergeRequiredNodeAffinities(
			userAffinity.RequiredDuringSchedulingIgnoredDuringExecution,                      // the user's node affinity
			(*generatedAffinity).NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, // our node affinity
		)
	}
}

// MergeRequiredNodeAffinities gets two required node affinity objects and returns a new object with merged affinities
// to maintain correctness of both node affinities, the idea is to  have all the permutations of one term from each node affinity.
// example:
//
// affinity1:
//
//	requiredDuringSchedulingIgnoredDuringExecution:
//	  nodeSelectorTerms:
//	  - termA
//	  - termB
//
// affinity2:
//
//	requiredDuringSchedulingIgnoredDuringExecution:
//	  nodeSelectorTerms:
//	  - termC
//	  - termD
//
// result:
//
// mergedAffinity:
//
//	requiredDuringSchedulingIgnoredDuringExecution:
//	  nodeSelectorTerms:
//	  - termA && termC
//	  - termA && termD
//	  - termB && termC
//	  - termB && termD
func MergeRequiredNodeAffinities(first, second *corev1.NodeSelector) *corev1.NodeSelector {
	if first == nil && second == nil {
		return nil
	} else if first == nil || len(first.NodeSelectorTerms) == 0 {
		return second
	} else if second == nil || len(second.NodeSelectorTerms) == 0 {
		return first
	}

	merged := &corev1.NodeSelector{
		NodeSelectorTerms: make([]corev1.NodeSelectorTerm, len(first.NodeSelectorTerms)*len(second.NodeSelectorTerms)),
	}

	i := 0
	for _, firstTerm := range first.NodeSelectorTerms {
		for _, secondTerm := range second.NodeSelectorTerms {
			merged.NodeSelectorTerms[i] = mergeAffinityTerms(firstTerm, secondTerm)
			i++
		}
	}

	return merged
}

// mergeAffinityTerms merges two terms by combining their match expressions (without duplicates)
func mergeAffinityTerms(first, second corev1.NodeSelectorTerm) corev1.NodeSelectorTerm {
	merged := corev1.NodeSelectorTerm{}
	alreadyMerged := make(map[string][]corev1.NodeSelectorRequirement)

	// loop over both terms
	for _, term := range []corev1.NodeSelectorTerm{first, second} {
	HandleExpr:
		for _, expr := range term.MatchExpressions {
			// as optimization there's no point in having duplicates in the merger,
			// check if the MatchExpression expr was already added to the merger
			if alreadyList, ok := alreadyMerged[expr.Key]; ok {
				for _, already := range alreadyList {
					if expr.Operator == already.Operator && utils.AreEquivalent(expr.Values, already.Values) {
						continue HandleExpr
					}
				}
			}

			// add expr to the merger
			merged.MatchExpressions = append(merged.MatchExpressions, expr)

			// mark expr as added
			alreadyMerged[expr.Key] = append(alreadyMerged[expr.Key], expr)
		}
	}

	return merged
}
