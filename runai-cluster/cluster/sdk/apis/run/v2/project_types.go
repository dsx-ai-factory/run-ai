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

package v2

import (
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/resourcing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProjectIdLabel = "runai/project-id"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	Department  string             `json:"department,omitempty"`
	Permissions ProjectPermissions `json:"permissions,omitempty"`

	Queues []QueueProjectSpec `json:"queues"`

	DefaultNodePools                             []string    `json:"defaultNodePools,omitempty"`
	InteractiveJobMaxIdleDurationSecs            int64       `json:"interactiveJobMaxIdleDurationSecs,omitempty"`
	InteractivePreemptibleJobMaxIdleDurationSecs int64       `json:"interactivePreemptibleJobMaxIdleDurationSecs,omitempty"`
	TrainingJobMaxIdleDurationSecs               int64       `json:"trainingJobMaxIdleDurationSecs,omitempty"`
	InteractiveJobTimeLimitSecs                  int64       `json:"interactiveJobTimeLimitSecs,omitempty"`
	TrainingJobTimeLimitSecs                     int64       `json:"trainingJobTimeLimitSecs,omitempty"`
	NodeAffinityInteractive                      []string    `json:"nodeAffinityInteractive,omitempty"`
	NodeAffinityTrain                            []string    `json:"nodeAffinityTrain,omitempty"`
	UpdatedAt                                    metav1.Time `json:"updatedAt,omitempty"`

	Namespace             string `json:"namespace,omitempty"`
	EnforceRunaiScheduler bool   `json:"enforceRunaiScheduler,omitempty"`

	DeletionRequest *DeletionRequest `json:"deletionRequest,omitempty"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	// The namespace of the project
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`

	// ProjectPhase is the recently observed phase of the project.
	Phase ProjectPhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ProjectPhase"`

	// A human-readable message indicating details about why the project is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// Current conditions of the project
	// +optional
	Conditions []ProjectCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`

	// Current situation of requested/allocated for each nodepool
	// +optional
	NodePoolsQuotaStatuses []QuotaStatus `json:"nodePoolsQuotaStatuses,omitempty" patchStrategy:"merge" patchMergeKey:"nodePoolName" protobuf:"bytes,2,rep,name=nodePoolsQuotaStatuses"`

	// Sum of quota statuses of all node pools
	// +optional
	QuotaStatus ProjectQuotaStatus `json:"quotaStatus,omitempty" protobuf:"bytes,3,opt,name=quotaStatus"`

	// +optional
	SyncStatus ProjectSyncStatus `json:"syncStatus,omitempty" protobuf:"bytes,3,opt,name=syncStatus"`
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

// TODO: after new department & queue crd, change Resources to be *QueueResources (this doesn't affect the crd itself)

type QueueProjectSpec struct {
	Name      string                     `json:"name"`
	Nodepool  string                     `json:"nodepool"`
	Resources *resourcing.QueueResources `json:"resources,omitempty"`
	Priority  *int32                     `json:"priority,omitempty"`
}

type DeletionType string

const (
	Orphan  DeletionType = "Orphan"
	Cascade DeletionType = "Cascade"
)

type DeletionRequest struct {
	Force bool         `json:"force,omitempty"`
	Type  DeletionType `json:"type,omitempty"`
}

type ProjectPermissions struct {
	Users  []string `json:"users,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

type ProjectPhase string

const (
	Ready    ProjectPhase = "Ready"
	Deleting ProjectPhase = "Deleting"
	NotReady ProjectPhase = "NotReady"
)

type ProjectCondition struct {
	Type ProjectConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=PodConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type ProjectConditionType string

// These are built-in conditions of project. An application may use a custom condition not listed here.
const (
	NamespaceReady      ProjectConditionType = "NamespaceReady"
	WorkloadsReady      ProjectConditionType = "WorkloadsReady"
	DataVolumesReady    ProjectConditionType = "DataVolumesReady"
	RoleBindingsReady   ProjectConditionType = "RoleBindingsReady"
	LimitRangeReady     ProjectConditionType = "LimitRangeReady"
	QueuesReady         ProjectConditionType = "QueuesReady"
	ServiceMonitorReady ProjectConditionType = "ServiceMonitorReady"
	SecretsReady        ProjectConditionType = "SecretsReady"
	PvcsReady           ProjectConditionType = "PvcsReady"
	ConfigMapsReady     ProjectConditionType = "ConfigMapsReady"
)

type QuotaStatus struct {
	// The NodePoolName of the current quota status
	NodePoolName string `json:"nodePoolName,omitempty" protobuf:"bytes,3,opt,name=nodePoolName"`

	// The detailed queue status
	QueueStatus kaiv2.QueueStatus `json:"queueStatus" protobuf:"bytes,2,opt,name=queueStatus,casttype=QueueStatus"`
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

type ProjectSyncStatus struct {
	// cluster-sync will set this value when syncing the object to control-plane
	// +optional
	LastSyncTimestamp metav1.Time `json:"lastSyncTimestamp,omitempty" protobuf:"bytes,1,opt,name=lastSyncTimestamp"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}

// IsOwner returns the index of this Project in the given object's OwnerReferences array if it exists.
func (project *Project) IsOwner(obj client.Object) (indexOfOwnerInRefArray int) {
	indexOfOwnerInRefArray = -1
	for index, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.Kind == "Project" &&
			ownerRef.UID == project.UID &&
			ownerRef.Name == project.Name {
			indexOfOwnerInRefArray = index
		}
	}
	return indexOfOwnerInRefArray
}
