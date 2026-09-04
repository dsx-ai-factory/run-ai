/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	nodepoolbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/nodepool"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (

	// DefaultAffinityWeight is the default weight for affinity
	// set for 10 so it will be low value - heuristically give the user's affinity higher weight while still allowing to provide lower value
	DefaultAffinityWeight = int32(10)
)

// AddPodAffinity adds pod affinity to make pods sit together on the same GPU rack
// Creates pod affinity for using its GPU network acceleration label as topology key
// Label key is "run.ai/top-owner-id"
// This lets the scheduler schedule pods on the same MNNVL "clique".
// The pod affinity is added as preferred
func AddPodAffinity(podAffinity *corev1.PodAffinity, gpuAccelerationLabel []string, workloadID string) {
	if podAffinity == nil {
		podAffinity = &corev1.PodAffinity{}
	}
	addedTopologyKeys := make(map[string]bool)
	for _, currentTerm := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		addedTopologyKeys[currentTerm.PodAffinityTerm.TopologyKey] = true
	}
	for _, accelerationLabel := range gpuAccelerationLabel {

		if accelerationLabel == "" || addedTopologyKeys[accelerationLabel] {
			continue
		}
		addedTopologyKeys[accelerationLabel] = true

		podAffinityTerm := corev1.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					commonconstants.TopOwnerUidLabel: workloadID,
				},
			},
			TopologyKey: accelerationLabel,
		}

		weightedPodAffinityTerm := corev1.WeightedPodAffinityTerm{
			Weight:          DefaultAffinityWeight,
			PodAffinityTerm: podAffinityTerm,
		}

		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			podAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			weightedPodAffinityTerm,
		)
	}
}

// NodeAffinityFromNodePools creates a node affinity with both the original nodepool labels
// and MNNVL-specific labels based on the provided mode
func NodeAffinityFromNodePools(nodePools []kaiv1alpha1.NodePool, affinity **corev1.Affinity, mode mnnvl.Mode) error {
	// Initialize affinity if nil
	initializeNodeAffinity(affinity)

	// Return error for ModeNone
	if !mode.IsEnabled() {
		return fmt.Errorf("MNNVL Node Affinity is not supported in None mode")
	}

	// Handle empty node pools
	if len(nodePools) == 0 {
		return nil
	}

	nodeAffinity := (*affinity).NodeAffinity

	// For Required mode: Add MNNVL requirements directly to node pool terms
	if mode == mnnvl.ModeRequired {
		addRequiredNodePoolAndMNNVLAffinity(nodePools, nodeAffinity)
		return nil
	}

	// For Preferred mode: Use the regular implementation then add MNNVL as preferred
	// Use regular node pool affinity implementation from common package
	if err := nodepoolbuilders.NodeAffinityFromNodePools(nodePools, affinity); err != nil {
		return err
	}

	// Add MNNVL as preferred
	addPreferredMNNVLAffinity(nodePools, (*affinity).NodeAffinity)

	return nil
}

// initializeNodeAffinity initializes the affinity and node affinity if they are nil
func initializeNodeAffinity(affinity **corev1.Affinity) {
	if *affinity == nil {
		*affinity = &corev1.Affinity{}
	}

	if (*affinity).NodeAffinity == nil {
		(*affinity).NodeAffinity = &corev1.NodeAffinity{}
	}
}

// addRequiredNodePoolAndMNNVLAffinity creates node pool affinity with MNNVL requirements
// included in the same terms (for Required mode)
func addRequiredNodePoolAndMNNVLAffinity(nodePools []kaiv1alpha1.NodePool, nodeAffinity *corev1.NodeAffinity) {
	terms := make([]corev1.NodeSelectorTerm, 0, len(nodePools))

	for _, nodePool := range nodePools {
		var expressions []corev1.NodeSelectorRequirement

		// 1. Add node pool requirement
		matchExpression := corev1.NodeSelectorRequirement{}

		nodepoolbuilders.CreateMatchExpressionForNodePool(&nodePool, &matchExpression)

		expressions = append(expressions, matchExpression)

		// 2. Add MNNVL requirement to the same term for ALL nodepools (as we assume all support MNNVL)
		mnnvlLabel := getGPUNetworkAccelerationLabel(&nodePool)
		if mnnvlLabel != "" {
			mnnvlRequirement := corev1.NodeSelectorRequirement{
				Key:      mnnvlLabel,
				Operator: corev1.NodeSelectorOpExists,
			}

			expressions = append(expressions, mnnvlRequirement)
		}
		// Create term with both requirements
		term := corev1.NodeSelectorTerm{
			MatchExpressions: expressions,
		}

		terms = append(terms, term)
	}

	// Set the required node affinity
	nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
		NodeSelectorTerms: terms,
	}
}

// addPreferredMNNVLAffinity adds preferred affinity terms for MNNVL
func addPreferredMNNVLAffinity(nodePools []kaiv1alpha1.NodePool, nodeAffinity *corev1.NodeAffinity) {

	// Track unique MNNVL label keys to avoid duplicates
	addedLabels := make(map[string]bool)

	// Add a preferred term for each node pool with unique MNNVL label
	for _, nodePool := range nodePools {
		label := getGPUNetworkAccelerationLabel(&nodePool)

		// Skip if label is empty or already added
		if label == "" || addedLabels[label] {
			continue
		}

		// Mark this label as added
		addedLabels[label] = true

		term := corev1.PreferredSchedulingTerm{
			Weight: DefaultAffinityWeight,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      label,
						Operator: corev1.NodeSelectorOpExists,
					},
				},
			},
		}

		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			term,
		)
	}
}

// getGPUNetworkAccelerationLabel returns the GPU network acceleration label for a nodepool.
// On the KAI CRD it is an annotation rather than a spec field, and it falls back to the
// default the legacy CRD schema applied, so an unset annotation behaves as before.
func getGPUNetworkAccelerationLabel(nodePool *kaiv1alpha1.NodePool) string {
	if labelKey := nodePool.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey]; labelKey != "" {
		return labelKey
	}
	return kaiv1alpha1.DefaultGPUNetworkAccelerationLabelKey
}
