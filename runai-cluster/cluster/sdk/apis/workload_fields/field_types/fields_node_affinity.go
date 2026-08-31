/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

import (
	corev1 "k8s.io/api/core/v1"
)

// +k8s:deepcopy-gen=true
type NodeAffinityFields struct {
	// If the affinity requirements specified by this field are not met at
	// scheduling time, the pod will not be scheduled onto the node.
	// If the affinity requirements specified by this field cease to be met
	// at some point during pod execution (e.g. due to an update), the system
	// may or may not try to eventually evict the pod from its node.
	NodeAffinityRequired *NodeSelector `json:"nodeAffinityRequired,omitempty"`
}

// A node selector represents the union of the results of one or more label queries
// over a set of nodes; that is, it represents the OR of the selectors represented
// by the node selector terms.
// +k8s:deepcopy-gen=true
type NodeSelector struct {
	// A list of node selector terms. The terms are ORed.
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

// NodeSelectorTerm A null or empty node selector term matches no objects.
// The requirements of them are ANDed.
// +k8s:deepcopy-gen=true
type NodeSelectorTerm struct {
	// A list of node selector requirements by node's labels.
	MatchExpressions []corev1.NodeSelectorRequirement `json:"matchExpressions,omitempty"`
}

func NewNodeAffinityFields(nodeAffinityRequiredTerms []NodeSelectorTerm) *NodeAffinityFields {
	return &NodeAffinityFields{
		NodeAffinityRequired: &NodeSelector{
			NodeSelectorTerms: nodeAffinityRequiredTerms,
		},
	}
}

// DerefRequired if initilized returns the NodeAffinityRequired selector, else returns an empty selector
func (na *NodeAffinityFields) DerefRequired() NodeSelector {
	if na == nil || na.NodeAffinityRequired == nil {
		return NodeSelector{}
	} else {
		return *na.NodeAffinityRequired
	}
}

// ToNodeAffinity converts the workload crd object to the equivalent k8s object
func (na *NodeAffinityFields) ToNodeAffinity() *corev1.NodeAffinity {
	if na == nil || na.NodeAffinityRequired == nil {
		return nil
	}

	convertedTerms := make([]corev1.NodeSelectorTerm, len(na.NodeAffinityRequired.NodeSelectorTerms))
	for i, term := range na.NodeAffinityRequired.NodeSelectorTerms {
		convertedTerms[i] = corev1.NodeSelectorTerm{
			MatchExpressions: term.MatchExpressions,
		}
	}

	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: convertedTerms,
		},
	}
}
