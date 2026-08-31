/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v2alpha1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	enforcer2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DistributedWorkload is a schema for the distributed workload API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="dw"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Based On",type=string,JSONPath=`.spec.baseWorkload`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DistributedWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this DistributedWorkload
	Spec dp.DistributedSpec `json:"spec,omitempty"`

	// The specifications of the master pod of this DistributedWorkload
	MasterSpec *dp.DistributedMasterSpec `json:"masterSpec,omitempty"`

	// The observed state of this DistributedWorkload
	Status DistributedWorkloadStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=true
type DistributedWorkloadStatus struct {
	common.WorkloadStatus       `json:",inline"`
	common.WorkloadDebugStatus  `json:",inline"`
	common.WorkloadCommonStatus `json:",inline"`
	common.NetworkStatus        `json:",inline"`
	// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
	SyncStatus *common.WorkloadSyncStatus `json:"syncStatus,omitempty"`
}

// DistributedWorkloadList contains a list of DistributedWorkload(s)
// +kubebuilder:object:root:=true
type DistributedWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributedWorkload `json:"items"`
}

func (self *DistributedWorkloadList) GetItems() []client.Object {
	objects := make([]client.Object, len(self.Items))
	for idx := 0; idx < len(self.Items); idx++ {
		objects[idx] = &self.Items[idx]
	}
	return objects
}

func (self *DistributedWorkloadList) GetWorkloadKind() string {
	return constants.KindDistributedWorkload
}

func init() {
	SchemeBuilder.Register(&DistributedWorkload{}, &DistributedWorkloadList{})
}

// functions to handle the workload status
func (dw *DistributedWorkload) Succeeded(mergeStack *enforcer2.MergeStack) *DistributedWorkload {
	dw.Status.CreationStatus = common.StatusSubmitted
	return dw
}

// NewFromJobWorkload copy generic job (interactive/training) to InteractiveWorkload
func NewFromJobWorkload(jobWorkload *jobs.JobWorkload) *DistributedWorkload {
	if jobWorkload != nil {
		result := &DistributedWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       constants.KindDistributedWorkload,
				APIVersion: GroupVersion.String(),
			},
		}

		jobWorkload.ObjectMeta.DeepCopyInto(&result.ObjectMeta)

		dp.LoadFromJobSpec(&jobWorkload.Spec, &result.Spec)

		if jobWorkload.MasterSpec != nil {
			result.MasterSpec = &dp.DistributedMasterSpec{}
			dp.LoadFromJobMasterSpec(jobWorkload.MasterSpec, result.MasterSpec)
		}

		result.Status = DistributedWorkloadStatus{
			WorkloadStatus:       jobWorkload.Status.WorkloadStatus,
			WorkloadDebugStatus:  jobWorkload.Status.WorkloadDebugStatus,
			WorkloadCommonStatus: jobWorkload.Status.WorkloadCommonStatus,
		}

		return result
	}

	return nil
}

func (self *DistributedWorkload) CopyTo(jobWorkload *jobs.JobWorkload) {

	if jobWorkload != nil {
		jobWorkload.TypeMeta = metav1.TypeMeta{
			Kind:       constants.KindDistributedWorkload,
			APIVersion: GroupVersion.String(),
		}
		self.ObjectMeta.DeepCopyInto(&jobWorkload.ObjectMeta)
		dp.CopyToJobSpec(&jobWorkload.Spec, &self.Spec)
		if self.MasterSpec != nil {
			if jobWorkload.MasterSpec == nil {
				jobWorkload.MasterSpec = &jobs.JobSpec{}
			}
			dp.CopyToJobMasterSpec(jobWorkload.MasterSpec, self.MasterSpec)
		}

		jobWorkload.Status = jobs.JobStatus{
			WorkloadStatus:       self.Status.WorkloadStatus,
			WorkloadDebugStatus:  self.Status.WorkloadDebugStatus,
			WorkloadCommonStatus: self.Status.WorkloadCommonStatus,
		}

	}
}

// GetChildName returns the name of the workload's children resources (e.g. run-ai jobs, services)
func (w *DistributedWorkload) GetChildName() string {
	name := w.ObjectMeta.Name
	// The job name should be metadata.Name, however for backward compatibility old workload uses sepc.Name
	// Therefor need to check if spec.Name isn't empty and use it instead of metadata.Name
	if w.Spec.Name != nil && w.Spec.Name.Deref() != "" {
		name = w.Spec.Name.Deref()
	}
	return name
}

func GetNewDistributedWorkloadFn() client.Object {
	return &DistributedWorkload{}
}
