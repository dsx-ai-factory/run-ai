/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrainingWorkload is a schema for the training workload API
// +kubebuilder:object:root:=true
// +kubebuilder:resource:shortName="tw"
// +kubebuilder:subresource:status
// +kubebuilder:unservedversion
// +kubebuilder:printcolumn:name="Based On",type=string,JSONPath=`.spec.baseWorkload`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.creationStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type TrainingWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//  The specifications of this TrainingWorkload
	Spec tpv1.TrainingSpec `json:"spec,omitempty"`

	//  The observed state of this TrainingWorkload
	Status v2alpha1.TrainingWorkloadStatus `json:"status,omitempty"`
}

// NewFromJobWorkload copy generic job (interactive/training) to training workload
func NewFromJobWorkload(jobWorkload *jobs.JobWorkload) *TrainingWorkload {

	if jobWorkload != nil {
		result := &TrainingWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       constants.KindTrainingWorkload,
				APIVersion: GroupVersion.String(),
			},
		}
		jobWorkload.ObjectMeta.DeepCopyInto(&result.ObjectMeta)
		tpv1.LoadFromJobSpec(&jobWorkload.Spec, &result.Spec)

		// copy the status
		result.Status.SyncStatus = jobWorkload.Status.SyncStatus
		result.Status.WorkloadCommonStatus = jobWorkload.Status.WorkloadCommonStatus
		result.Status.WorkloadDebugStatus = jobWorkload.Status.WorkloadDebugStatus
		result.Status.WorkloadStatus = jobWorkload.Status.WorkloadStatus
		result.Status.NetworkStatus = jobWorkload.Status.NetworkStatus
		return result
	}
	return nil
}

func (self *TrainingWorkload) CopyTo(jobWorkload *jobs.JobWorkload) {

	if jobWorkload != nil {
		jobWorkload.TypeMeta = metav1.TypeMeta{
			Kind:       constants.KindTrainingWorkload,
			APIVersion: GroupVersion.String(),
		}
		self.ObjectMeta.DeepCopyInto(&jobWorkload.ObjectMeta)
		self.Spec.CopyTo(&jobWorkload.Spec)

		// copy the status
		jobWorkload.Status.SyncStatus = self.Status.SyncStatus
		jobWorkload.Status.WorkloadCommonStatus = self.Status.WorkloadCommonStatus
		jobWorkload.Status.WorkloadDebugStatus = self.Status.WorkloadDebugStatus
		jobWorkload.Status.WorkloadStatus = self.Status.WorkloadStatus
		jobWorkload.Status.NetworkStatus = self.Status.NetworkStatus

	}
}

// TrainingWorkloadList contains a list of TraiingWorkload(s)
// +kubebuilder:object:root:=true
type TrainingWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrainingWorkload `json:"items"`
}

func (self *TrainingWorkloadList) AsJobWorkLoad() []jobs.JobWorkload {

	result := make([]jobs.JobWorkload, len(self.Items))
	for idx := 0; idx < len(self.Items); idx++ {
		self.Items[idx].CopyTo(&result[idx])
	}
	return result
}

func init() {
	SchemeBuilder.Register(&TrainingWorkload{}, &TrainingWorkloadList{})
}
