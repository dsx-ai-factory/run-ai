/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodeexclusion

import (
	"slices"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
)

const hostnameKey = corev1.LabelHostname

// MergeNodeExclusions merges node exclusions into existing NodeAffinityFields.
// If existing is nil or has no NodeAffinityRequired, creates a new one with the exclusions.
func MergeNodeExclusions(existing *field_types.NodeAffinityFields, exclusions []string) *field_types.NodeAffinityFields {
	if len(exclusions) == 0 {
		return existing
	}

	result := existing
	if result == nil || result.NodeAffinityRequired == nil {
		result = &field_types.NodeAffinityFields{
			NodeAffinityRequired: &field_types.NodeSelector{},
		}
	}

	terms := result.NodeAffinityRequired.NodeSelectorTerms
	// Ensure at least one term exists so exclusions have somewhere to attach.
	if len(terms) == 0 {
		terms = []field_types.NodeSelectorTerm{{}}
	}
	for i := range terms {
		addExclusionToTerm(&terms[i], exclusions)
	}
	result.NodeAffinityRequired.NodeSelectorTerms = terms

	return result
}

func addExclusionToTerm(term *field_types.NodeSelectorTerm, exclusions []string) {
	for i := range term.MatchExpressions {
		if term.MatchExpressions[i].Key == hostnameKey && term.MatchExpressions[i].Operator == corev1.NodeSelectorOpNotIn {
			term.MatchExpressions[i].Values = unionStrings(term.MatchExpressions[i].Values, exclusions)
			return
		}
	}
	term.MatchExpressions = append(term.MatchExpressions, corev1.NodeSelectorRequirement{
		Key:      hostnameKey,
		Operator: corev1.NodeSelectorOpNotIn,
		Values:   unionStrings(nil, exclusions),
	})
}

func unionStrings(a, b []string) []string {
	result := make([]string, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)
	slices.Sort(result)
	return slices.Compact(result)
}
