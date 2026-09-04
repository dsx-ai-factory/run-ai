/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v2alpha1

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DistributedPolicy is a schema for the distributed policy API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="dp"
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DistributedPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this DistributedPolicy
	Spec DistributedSpec `json:"spec,omitempty"`

	// The specifications for the master pod of this DistributedPolicy
	MasterSpec *DistributedMasterSpec `json:"masterSpec,omitempty"`

	// The observed state of this DistributedPolicy
	Status common.PolicyStatus `json:"status,omitempty"`
}

type DistributedSpec struct {
	field_groups.WorkloadFields           `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.AccessControlFields      `json:""`
	field_groups.DistributedFields        `json:""`
	field_groups.StateFields              `json:""`
	field_groups.SchedulingFields         `json:""`
}

type DistributedMasterSpec struct {
	field_groups.WorkloadFields           `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.AccessControlFields      `json:""`
}

// DistributedPolicyList contains a list of DistributedPolicy
// +kubebuilder:object:root:=true
type DistributedPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributedPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistributedPolicy{}, &DistributedPolicyList{})
}

// NewEmptyDistributedPolicy - we need empty policy for the case that the administrator does not provide policy
// in this case, validation will only consider the run:ai global rules
func NewEmptyDistributedPolicy() *DistributedPolicy {
	return &DistributedPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
		},
		Spec: DistributedSpec{},
	}
}

// LoadFromJobSpec copy generic job to distributed spec
func LoadFromJobSpec(jobSpec *jobs.JobSpec, spec *DistributedSpec) {
	if spec != nil && jobSpec != nil {
		jobSpec.WorkloadFields.DeepCopyInto(&spec.WorkloadFields)
		jobSpec.ResourceAllocationFields.DeepCopyInto(&spec.ResourceAllocationFields)
		jobSpec.RunaiSchedulerFields.DeepCopyInto(&spec.RunaiSchedulerFields)
		jobSpec.AccessControlFields.DeepCopyInto(&spec.AccessControlFields)
		jobSpec.DistributedFields.DeepCopyInto(&spec.DistributedFields)
		jobSpec.StateFields.DeepCopyInto(&spec.StateFields)
		jobSpec.SchedulingFields.DeepCopyInto(&spec.SchedulingFields)
	}
}

// LoadFromJobMasterSpec copy generic job to distributed master spec
func LoadFromJobMasterSpec(jobMasterSpec *jobs.JobSpec, masterSpec *DistributedMasterSpec) {
	if masterSpec != nil && jobMasterSpec != nil {
		jobMasterSpec.WorkloadFields.DeepCopyInto(&masterSpec.WorkloadFields)
		jobMasterSpec.ResourceAllocationFields.DeepCopyInto(&masterSpec.ResourceAllocationFields)
		jobMasterSpec.RunaiSchedulerFields.DeepCopyInto(&masterSpec.RunaiSchedulerFields)
		jobMasterSpec.AccessControlFields.DeepCopyInto(&masterSpec.AccessControlFields)
	}
}

// CopyToJobSpec copy generic job to distributed spec
func CopyToJobSpec(jobSpec *jobs.JobSpec, spec *DistributedSpec) {
	if spec != nil && jobSpec != nil {
		spec.WorkloadFields.DeepCopyInto(&jobSpec.WorkloadFields)
		spec.ResourceAllocationFields.DeepCopyInto(&jobSpec.ResourceAllocationFields)
		spec.RunaiSchedulerFields.DeepCopyInto(&jobSpec.RunaiSchedulerFields)
		spec.AccessControlFields.DeepCopyInto(&jobSpec.AccessControlFields)
		spec.DistributedFields.DeepCopyInto(&jobSpec.DistributedFields)
		spec.StateFields.DeepCopyInto(&jobSpec.StateFields)
		spec.SchedulingFields.DeepCopyInto(&jobSpec.SchedulingFields)
	}
}

// CopyToJobMasterSpec copy generic job to distributed spec
func CopyToJobMasterSpec(jobMasterSpec *jobs.JobSpec, masterSpec *DistributedMasterSpec) {
	if masterSpec != nil && jobMasterSpec != nil {
		masterSpec.WorkloadFields.DeepCopyInto(&jobMasterSpec.WorkloadFields)
		masterSpec.ResourceAllocationFields.DeepCopyInto(&jobMasterSpec.ResourceAllocationFields)
		masterSpec.RunaiSchedulerFields.DeepCopyInto(&jobMasterSpec.RunaiSchedulerFields)
		masterSpec.AccessControlFields.DeepCopyInto(&jobMasterSpec.AccessControlFields)
	}
}
