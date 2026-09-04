/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Schema for inference policy API
// +kubebuilder:object:root:=true
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName="infp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
	field_groups.WorkloadFields           `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.InferenceFields          `json:""`
	field_groups.SchedulingFields         `json:""`
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

func inferenceWorkloadRules(genRules field_groups.WorkloadFields) field_groups.WorkloadFields {
	port := genRules.NetworkingFields.Ports
	if port.ItemRules.Protocol == nil {
		port.ItemRules.Protocol = &flds.StringRules{}
	}

	port.ItemRules.Protocol.Options = &[]flds.OptionRule{
		{Value: ptr.To(constants.ProtocolGRPC), Name: ptr.To("gRPC")},
		{Value: ptr.To(constants.ProtocolHTTP), Name: ptr.To("HTTP")},
	}

	if port.ItemRules.ContainerPort == nil {
		port.ItemRules.ContainerPort = &flds.IntRules{
			Min: ptr.To(0),
			Max: ptr.To(65535),
		}
	}
	port.ItemRules.ContainerPort.Required = ptr.To(true)
	port.ItemRules.ContainerPort.Format = ptr.To(workload_fields.FormatPort)

	return genRules
}

// LoadFromJobSpec copy generic job to inference spec
func LoadFromJobSpec(jobSpec *jobs.JobSpec, spec *InferenceSpec) {
	if spec != nil && jobSpec != nil {
		jobSpec.WorkloadFields.DeepCopyInto(&spec.WorkloadFields)
		jobSpec.ResourceAllocationFields.DeepCopyInto(&spec.ResourceAllocationFields)
		jobSpec.RunaiSchedulerFields.DeepCopyInto(&spec.RunaiSchedulerFields)
		jobSpec.InferenceFields.DeepCopyInto(&spec.InferenceFields)
		jobSpec.SchedulingFields.DeepCopyInto(&spec.SchedulingFields)
	}
}

func (self *InferenceSpec) CopyTo(jobSpec *jobs.JobSpec) {
	if jobSpec != nil {
		self.WorkloadFields.DeepCopyInto(&jobSpec.WorkloadFields)
		self.ResourceAllocationFields.DeepCopyInto(&jobSpec.ResourceAllocationFields)
		self.RunaiSchedulerFields.DeepCopyInto(&jobSpec.RunaiSchedulerFields)
		self.InferenceFields.DeepCopyInto(&jobSpec.InferenceFields)
		self.SchedulingFields.DeepCopyInto(&jobSpec.SchedulingFields)
	}
}
