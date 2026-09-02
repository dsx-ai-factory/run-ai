/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package v1alpha1

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root:=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas.value,statuspath=.status.replicas
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName="dinfw"
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="url",type="string",JSONPath=".status.externalUrls[-1]"
// Schema for distributed inference workload API
type DistributedInferenceWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// The specifications of this workload
	Spec dp.DistributedInferenceSpec `json:"spec,omitempty"`

	// The specifications of the leader pod of this DistributedInferenceWorkload
	LeaderSpec dp.DistributedInferenceEntitySpec `json:"leaderSpec,omitempty"`

	// The specifications of the worker pods of this DistributedInferenceWorkload
	WorkerSpec *dp.DistributedInferenceEntitySpec `json:"workerSpec,omitempty"`

	// The observed state of this workload
	Status DistributedInferenceWorkloadStatus `json:"status,omitempty"`
}

type DistributedInferenceWorkloadStatus struct {
	common.WorkloadDebugStatus        `json:""`
	common.NetworkStatus              `json:""`
	common.DistributedInferenceStatus `json:""`
	// WorkloadSyncStatus defines the status of the sync, reported by cluster-sync
	SyncStatus *common.WorkloadSyncStatus `json:"syncStatus,omitempty"`
}

// +kubebuilder:object:root:=true
// DistributedInferenceWorkloadList contains a list of DistributedInferenceWorkload(s)
type DistributedInferenceWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributedInferenceWorkload `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DistributedInferenceWorkload{}, &DistributedInferenceWorkloadList{})
}

// GetChildName returns the name of the workload's children resources (e.g. run-ai jobs, services)
func (dinfw *DistributedInferenceWorkload) GetChildName() string {
	return dinfw.ObjectMeta.Name
}
