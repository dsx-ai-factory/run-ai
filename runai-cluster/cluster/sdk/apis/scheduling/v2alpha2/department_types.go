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

package v2alpha2

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DepartmentIdLabel   = "runai/department-id"
	DepartmentNameLabel = "runai/department-name"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DepartmentSpec defines the desired state of Department
type DepartmentSpec struct {
	DeservedGPUs     float64                    `json:"deservedGpus,omitempty" protobuf:"bytes,1,opt,name=deservedGpus"`
	MaxAllowedGPUs   *float64                   `json:"maxAllowedGpus,omitempty" protobuf:"bytes,1,opt,name=maxAllowedGpus"`
	Resources        *resourcing.QueueResources `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
	Queues           []QueueDepartmentSpec      `json:"queues"`
	DepartmentAdmins []string                   `json:"departmentAdmins,omitempty"`
	UpdatedAt        metav1.Time                `json:"updatedAt,omitempty"`
}

type QueueDepartmentSpec struct {
	Name                        string                     `json:"name"`
	Nodepool                    string                     `json:"nodepool"`
	InteractiveJobTimeLimitSecs int64                      `json:"interactiveJobTimeLimitSecs,omitempty"`
	NodeAffinityInteractive     []string                   `json:"nodeAffinityInteractive,omitempty"`
	NodeAffinityTrain           []string                   `json:"nodeAffinityTrain,omitempty"`
	Resources                   *resourcing.QueueResources `json:"resources,omitempty"`
	SwapEnabled                 bool                       `json:"swapEnabled,omitempty"`
	Priority                    *int32                     `json:"priority,omitempty"`
}

// DepartmentStatus defines the observed state of Department
type DepartmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	SyncStatus DepartmentSyncStatus `json:"syncStatus,omitempty" protobuf:"bytes,1,opt,name=syncStatus"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// Department is the Schema for the departments API
type Department struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DepartmentSpec   `json:"spec,omitempty"`
	Status DepartmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DepartmentList contains a list of Department
type DepartmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Department `json:"items"`
}

type DepartmentSyncStatus struct {
	// cluster-sync will set this value when syncing the object to control-plane
	// +optional
	LastSyncTimestamp metav1.Time `json:"lastSyncTimestamp,omitempty" protobuf:"bytes,1,opt,name=lastSyncTimestamp"`
}

func init() {
	SchemeBuilder.Register(&Department{}, &DepartmentList{})
}

func GetDepartmentId(objLabels map[string]string) string {
	if depId, found := objLabels[DepartmentIdLabel]; found {
		return depId
	}
	return ""
}
