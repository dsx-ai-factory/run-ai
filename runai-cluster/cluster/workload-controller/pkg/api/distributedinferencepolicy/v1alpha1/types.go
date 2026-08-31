/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1alpha1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Schema for distributed inference policy API
// +kubebuilder:object:root:=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName="dinfp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DistributedInferencePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this DistributedInferencePolicy
	Spec DistributedInferenceSpec `json:"spec,omitempty"`

	// The specifications of the leader pod of this DistributedInferencePolicy
	LeaderSpec DistributedInferenceEntitySpec `json:"leaderSpec,omitempty"`

	// The specifications of the worker pods of this DistributedInferencePolicy
	WorkerSpec *DistributedInferenceEntitySpec `json:"workerSpec,omitempty"`

	// The observed state of this DistributedInferencePolicy
	Status common.PolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true
type DistributedInferenceSpec struct {
	field_groups.IdentityFields                      `json:""`
	field_groups.DistributedInferenceFields          `json:""`
	field_groups.SchedulingFields                    `json:""`
	field_groups.DistributedInferenceSchedulerFields `json:""`
	field_groups.SecurityFields                      `json:""`
}

// +k8s:deepcopy-gen=true
type DistributedInferenceEntitySpec struct {
	field_groups.ContainerFields                           `json:""`
	field_groups.StorageFields                             `json:""`
	field_groups.ResourceAllocationFields                  `json:""`
	field_groups.DistributedInferenceEntitySchedulerFields `json:""`
	field_groups.SecurityFields                            `json:""`
}

// +kubebuilder:object:root:=true
// DistributedInferencePolicyList contains a list of DistributedInferencePolicy
type DistributedInferencePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributedInferencePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistributedInferencePolicy{}, &DistributedInferencePolicyList{})
}

// we need empty policy when there is no policy for a given type.
// it is not mandatory for administrator to provided rules, and in this case
// validation will only consider the run:ai global rules
func NewEmptyDistributedInferencePolicy() *DistributedInferencePolicy {
	return &DistributedInferencePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
		},

		Spec:       DistributedInferenceSpec{},
		LeaderSpec: DistributedInferenceEntitySpec{},
		WorkerSpec: &DistributedInferenceEntitySpec{},
	}
}
