/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InteractivePolicy is a schema for the interactive policy API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="ip"
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type InteractivePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this InteractivePolicy
	Spec InteractiveSpec `json:"spec,omitempty"`

	// The observed state of this InteractivePolicy
	Status common.PolicyStatus `json:"status,omitempty"`
}

type InteractiveSpec struct {
	field_groups.WorkloadFields           `json:""`
	field_groups.MpiFields                `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.AccessControlFields      `json:""`
	field_groups.StateFields              `json:""`
	field_groups.InteractiveFields        `json:""`
	field_groups.RunaiJobSpecificRules    `json:""`
	field_groups.SchedulingFields         `json:""`
}

// LoadFromJobSpec copy generic job (interactive/training) to interactive spec
func LoadFromJobSpec(jobSpec *jobs.JobSpec, spec *InteractiveSpec) {
	if jobSpec != nil {
		jobSpec.WorkloadFields.DeepCopyInto(&spec.WorkloadFields)
		jobSpec.MpiFields.DeepCopyInto(&spec.MpiFields)
		jobSpec.ResourceAllocationFields.DeepCopyInto(&spec.ResourceAllocationFields)
		jobSpec.RunaiSchedulerFields.DeepCopyInto(&spec.RunaiSchedulerFields)
		jobSpec.AccessControlFields.DeepCopyInto(&spec.AccessControlFields)
		jobSpec.InteractiveFields.DeepCopyInto(&spec.InteractiveFields)
		jobSpec.StateFields.DeepCopyInto(&spec.StateFields)
		jobSpec.RunaiJobSpecificRules.DeepCopyInto(&spec.RunaiJobSpecificRules)
		jobSpec.SchedulingFields.DeepCopyInto(&spec.SchedulingFields)
	}
}

func (self *InteractiveSpec) CopyTo(jobSpec *jobs.JobSpec) {
	if jobSpec != nil {
		self.WorkloadFields.DeepCopyInto(&jobSpec.WorkloadFields)
		self.MpiFields.DeepCopyInto(&jobSpec.MpiFields)
		self.ResourceAllocationFields.DeepCopyInto(&jobSpec.ResourceAllocationFields)
		self.RunaiSchedulerFields.DeepCopyInto(&jobSpec.RunaiSchedulerFields)
		self.AccessControlFields.DeepCopyInto(&jobSpec.AccessControlFields)
		self.InteractiveFields.DeepCopyInto(&jobSpec.InteractiveFields)
		self.StateFields.DeepCopyInto(&jobSpec.StateFields)
		self.RunaiJobSpecificRules.DeepCopyInto(&jobSpec.RunaiJobSpecificRules)
		self.SchedulingFields.DeepCopyInto(&jobSpec.SchedulingFields)
	}
}

// InteractivePolicyList contains a list of InteractivePolicy
// +kubebuilder:object:root:=true
type InteractivePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InteractivePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InteractivePolicy{}, &InteractivePolicyList{})
}

// NewEmptyInteractivePolicy - we need empty policy for the case that the administrator does not provide policy
// in this case, validation will only consider the run:ai global rules
func NewEmptyInteractivePolicy() *InteractivePolicy {
	return &InteractivePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
		},
		Spec: InteractiveSpec{},
	}
}
