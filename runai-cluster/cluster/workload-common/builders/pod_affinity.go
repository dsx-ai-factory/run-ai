/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UsersAffinityWeight = 100
)

func CreatePodAffinity(PodAffinity bool, PodAffinityTopology string, PodAffinitySchedulingRule string, affinity **corev1.Affinity, jobName string, isDistributed bool) error {

	if affinity == nil || !PodAffinity {
		return nil
	}

	if PodAffinitySchedulingRule == "" {
		PodAffinitySchedulingRule = constants.RequiredDuringSchedulingIgnoredDuringExecutionValue
	}

	labelSelectorKey := constants.LabelRelease
	if isDistributed {
		labelSelectorKey = constants.LabelTrainingKubeFlow
	}
	// the pod label selector for the pod affinity
	labelSelector := &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      labelSelectorKey,
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{jobName},
			},
		},
	}

	var podAffinityPreferredArray []corev1.WeightedPodAffinityTerm
	var podAffinityRequiredArray []corev1.PodAffinityTerm

	if PodAffinitySchedulingRule == constants.RequiredDuringSchedulingIgnoredDuringExecutionValue {
		podAffinityRequiredArray = []corev1.PodAffinityTerm{
			{
				TopologyKey:   PodAffinityTopology,
				LabelSelector: labelSelector,
			},
		}
	} else if PodAffinitySchedulingRule == constants.PreferredDuringSchedulingIgnoredDuringExecutionValue {
		podAffinityPreferredArray = []corev1.WeightedPodAffinityTerm{
			{
				Weight: UsersAffinityWeight,
				PodAffinityTerm: corev1.PodAffinityTerm{
					TopologyKey:   PodAffinityTopology,
					LabelSelector: labelSelector,
				},
			},
		}
	} else {
		// BUG: Returning nil in case of an invalid PodAffinitySchedulingRule.
		// This should return an error instead.
		return nil
	}

	// when using affinity should be careful to not override existing affinity
	// both pod affinity and node affinity are part of affinity
	if (*affinity) == nil {
		*affinity = &corev1.Affinity{
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  podAffinityRequiredArray,
				PreferredDuringSchedulingIgnoredDuringExecution: podAffinityPreferredArray,
			},
		}
	} else {
		if (*affinity).PodAffinity == nil {
			(*affinity).PodAffinity = &corev1.PodAffinity{}
		}
		(*affinity).PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			(*affinity).PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			podAffinityRequiredArray...,
		)
		(*affinity).PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			(*affinity).PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			podAffinityPreferredArray...,
		)
	}

	return nil
}
