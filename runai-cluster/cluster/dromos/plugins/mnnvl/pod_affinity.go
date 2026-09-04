/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodAffinityWeight is the weight on the grouping affinity's preferred terms. Kept
// low so user-supplied affinities outrank it.
const PodAffinityWeight = int32(10)

// BuildGroupingAffinity builds a preferred pod affinity that co-locates pods sharing
// selectorLabel=value, one term per topology key. Returns nil when there are no usable
// topology keys. Shared by the plugin's default grouping and consumer providers.
func BuildGroupingAffinity(selectorLabel, value string, topologyKeys []string) *corev1.PodAffinity {
	terms := make([]corev1.WeightedPodAffinityTerm, 0, len(topologyKeys))
	for _, key := range topologyKeys {
		if key == "" {
			continue
		}
		terms = append(terms, corev1.WeightedPodAffinityTerm{
			Weight: PodAffinityWeight,
			PodAffinityTerm: corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{selectorLabel: value},
				},
				TopologyKey: key,
			},
		})
	}
	if len(terms) == 0 {
		return nil
	}
	return &corev1.PodAffinity{PreferredDuringSchedulingIgnoredDuringExecution: terms}
}

// mergePodAffinity appends src's terms into dst, skipping any whose TopologyKey is
// already there. Dedup keeps repeated reconciles convergent and preserves user terms.
func mergePodAffinity(dst, src *corev1.PodAffinity) {
	if dst == nil || src == nil {
		return
	}

	preferredKeys := make(map[string]bool)
	for _, term := range dst.PreferredDuringSchedulingIgnoredDuringExecution {
		preferredKeys[term.PodAffinityTerm.TopologyKey] = true
	}
	for _, term := range src.PreferredDuringSchedulingIgnoredDuringExecution {
		if preferredKeys[term.PodAffinityTerm.TopologyKey] {
			continue
		}
		preferredKeys[term.PodAffinityTerm.TopologyKey] = true
		dst.PreferredDuringSchedulingIgnoredDuringExecution = append(
			dst.PreferredDuringSchedulingIgnoredDuringExecution, term)
	}

	requiredKeys := make(map[string]bool)
	for _, term := range dst.RequiredDuringSchedulingIgnoredDuringExecution {
		requiredKeys[term.TopologyKey] = true
	}
	for _, term := range src.RequiredDuringSchedulingIgnoredDuringExecution {
		if requiredKeys[term.TopologyKey] {
			continue
		}
		requiredKeys[term.TopologyKey] = true
		dst.RequiredDuringSchedulingIgnoredDuringExecution = append(
			dst.RequiredDuringSchedulingIgnoredDuringExecution, term)
	}
}
