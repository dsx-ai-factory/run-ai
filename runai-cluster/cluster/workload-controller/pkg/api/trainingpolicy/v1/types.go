/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrainingPolicy is a schema for training jobs policy API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="tp"
// +kubebuilder:unservedversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type TrainingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this TrainingPolicy
	Spec TrainingSpec `json:"spec,omitempty"`

	// The observed state of this TrainingPolicy
	Status common.PolicyStatus `json:"status,omitempty"`
}

type TrainingSpec struct {
	field_groups.WorkloadFields `json:""`

	field_groups.MpiFields                `json:""`
	field_groups.LifeCycleFields          `json:""`
	field_groups.ResourceAllocationFields `json:""`
	field_groups.RunaiSchedulerFields     `json:""`
	field_groups.AccessControlFields      `json:""`
	field_groups.RunaiJobSpecificRules    `json:""`
}

// LoadFromJobSpec copy from generic job (interactive/training) to TrainingSpec
func LoadFromJobSpec(jobSpec *jobs.JobSpec, spec *TrainingSpec) {
	if jobSpec != nil {
		jobSpec.WorkloadFields.DeepCopyInto(&spec.WorkloadFields)
		jobSpec.MpiFields.DeepCopyInto(&spec.MpiFields)
		jobSpec.LifeCycleFields.DeepCopyInto(&spec.LifeCycleFields)
		jobSpec.ResourceAllocationFields.DeepCopyInto(&spec.ResourceAllocationFields)
		jobSpec.RunaiSchedulerFields.DeepCopyInto(&spec.RunaiSchedulerFields)
		jobSpec.AccessControlFields.DeepCopyInto(&spec.AccessControlFields)
		jobSpec.RunaiJobSpecificRules.DeepCopyInto(&spec.RunaiJobSpecificRules)
	}
}

func (self *TrainingSpec) CopyTo(jobSpec *jobs.JobSpec) {
	if jobSpec != nil {
		self.WorkloadFields.DeepCopyInto(&jobSpec.WorkloadFields)
		self.MpiFields.DeepCopyInto(&jobSpec.MpiFields)
		self.LifeCycleFields.DeepCopyInto(&jobSpec.LifeCycleFields)
		self.ResourceAllocationFields.DeepCopyInto(&jobSpec.ResourceAllocationFields)
		self.RunaiSchedulerFields.DeepCopyInto(&jobSpec.RunaiSchedulerFields)
		self.AccessControlFields.DeepCopyInto(&jobSpec.AccessControlFields)
		self.RunaiJobSpecificRules.DeepCopyInto(&jobSpec.RunaiJobSpecificRules)
	}
}

// TrainingPolicyList contains a list of TrainingPolicy
// +kubebuilder:object:root:=true
type TrainingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrainingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrainingPolicy{}, &TrainingPolicyList{})
}

// NewEmptyTrainingPolicy - we need empty policy for the case that the administrator does not provide policy
// in this case, validation will only consider the run:ai global rules
func NewEmptyTrainingPolicy() *TrainingPolicy {
	return &TrainingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.NamespaceRunAi,
		},
		Spec: TrainingSpec{},
	}
}
