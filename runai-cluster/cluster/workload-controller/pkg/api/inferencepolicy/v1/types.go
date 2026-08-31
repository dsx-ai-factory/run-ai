/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root:=true
// +kubebuilder:subresource:status
// +kubebuilder:unservedversion
// +kubebuilder:resource:shortName="infp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// Schema for inference policy API
type InferencePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this InferencePolicy
	Spec InferenceSpec `json:"spec,omitempty"`

	// The observed state of this InferencePolicy
	Status common.PolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true
type InferenceSpec struct {
	field_groups.WorkloadFields `json:""`

	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`

	field_groups.AutoScalingFields `json:""`
}

// +kubebuilder:object:root:=true
// InferencePolicyList contains a list of InferencePolicy
type InferencePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferencePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferencePolicy{}, &InferencePolicyList{})
}

// we need empty policy when there is no policy for a given type.
// it is not mandatory for administrator to provided rules, and in this case
// validation will only consider the run:ai global rules
func NewEmptyInferencePolicy() *InferencePolicy {
	return &InferencePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
		},

		Spec: InferenceSpec{},
	}
}

func (self *InferenceSpec) CopyTo(jobSpec *jobs.JobSpec) {
	if jobSpec != nil {
		self.WorkloadFields.DeepCopyInto(&jobSpec.WorkloadFields)
		self.ResourceAllocationFields.DeepCopyInto(&jobSpec.ResourceAllocationFields)
		self.RunaiSchedulerFields.DeepCopyInto(&jobSpec.RunaiSchedulerFields)
		self.AutoScalingFields.DeepCopyInto(&jobSpec.AutoScalingFields)
	}
}
