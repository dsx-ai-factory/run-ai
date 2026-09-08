/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedNodesConfigSpec defines the desired state of the managed nodes config
type ManagedNodesConfigSpec struct {
	// Node selector for unmanaged nodes
	InclusionCriteria v1.NodeSelector `json:"inclusion_criteria,omitempty"`
}

// ManagedNodesConfigStatus defines the observed state of ManagedNodesConfig
type ManagedNodesConfigStatus struct {
	// Conditions Defines the observed state of the CRD in the cluster
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration specifies the latest generation of CRD
	// the operator was aware of
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// ManagedNodesConfig is the Schema for the managed nodes config API
type ManagedNodesConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedNodesConfigSpec   `json:"spec,omitempty"`
	Status ManagedNodesConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedNodesConfigList contains a list of ManagedNodesConfig
type ManagedNodesConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedNodesConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedNodesConfig{}, &ManagedNodesConfigList{})
}

type MNCConditionType string

// These are built-in conditions of managed nodes config. An application may use a custom condition not listed here.
const (
	MNCConditionTypeApplied MNCConditionType = "Applied"
)

type MNCConditionReason string

const (
	MNCConditionReasonAllNodesIncludedCorrectly MNCConditionReason = "AllNodesIncludedCorrectly"
	MNCConditionReasonToBeExcludedNodes         MNCConditionReason = "ToBeExcludedNodes"
)
