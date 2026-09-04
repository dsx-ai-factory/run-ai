/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// NamespaceProjectLabelKey is the label key written on a project's namespace (and read back to
	// resolve a namespace's project) to identify the project it belongs to.
	NamespaceProjectLabelKey = "kai/project"
)

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	// The department parent of the project
	// +optional
	Parent string `json:"parent,omitempty"`

	// Project's queues
	Queues []QueueConfig `json:"queues"`

	// Ordered list of node pools that defines the default scheduling preference for workloads in this project
	DefaultNodePools []string `json:"defaultNodePools,omitempty"`

	// Requested namespace associated with the project. If not specified it will be generated from the project name
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Whether to run KAI scheduler on the project's workloads
	EnforceKaiScheduler bool `json:"enforceKaiScheduler,omitempty"`

	// Requested deletion type for the project
	DeletionType *DeletionType `json:"deletionType,omitempty"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	// Actual namespace of the project
	Namespace string `json:"namespace,omitempty"`

	// ProjectPhase is the recently observed phase of the project.
	Phase ProjectPhase `json:"phase,omitempty"`

	// A human-readable message indicating details about why the project is in this condition.
	// +optional
	Message string `json:"message,omitempty"`

	// Current conditions of the project
	// +optional
	Conditions []ProjectCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Current situation of requested/allocated for each nodepool
	// +optional
	NodePoolsQuotaStatuses []QuotaStatus `json:"nodePoolsQuotaStatuses,omitempty" patchStrategy:"merge" patchMergeKey:"nodePoolName"`

	// Sum of quota statuses of all node pools
	// +optional
	QuotaStatus ProjectQuotaStatus `json:"quotaStatus,omitempty"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

type DeletionType string

const (
	Blocking DeletionType = "Blocking"
)

type ProjectPhase string

const (
	Ready    ProjectPhase = "Ready"
	NotReady ProjectPhase = "NotReady"
)

type ProjectCondition struct {
	Type ProjectConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type ProjectConditionType string

// These are built-in conditions of project. An application may use a custom condition not listed here.
const (
	NamespaceReady    ProjectConditionType = "NamespaceReady"
	RoleBindingsReady ProjectConditionType = "RoleBindingsReady"
	QueuesReady       ProjectConditionType = "QueuesReady"
)

type QuotaStatus struct {
	// The NodePoolName of the current quota status
	NodePoolName string `json:"nodePoolName,omitempty"`

	// The detailed queue status
	QueueStatus schedv2.QueueStatus `json:"queueStatus"`
}

// ProjectQuotaStatus defines the sum of all Queues quota statuses
type ProjectQuotaStatus struct {
	// Current allocated GPU (in fractions), CPU (in millicpus) and Memory in megabytes
	// for all running jobs in queue and child queues
	Allocated corev1.ResourceList `json:"allocated,omitempty"`

	// Current allocated GPU (in fractions), CPU (in millicpus) and Memory in megabytes
	// for all non-preemptible running jobs in queue and child queues
	AllocatedNonPreemptible corev1.ResourceList `json:"allocatedNonPreemptible,omitempty"`

	// Current requested GPU (in fractions), CPU (in millicpus) and Memory in megabytes
	// by all running and pending jobs in queue and child queues
	Requested corev1.ResourceList `json:"requested,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
