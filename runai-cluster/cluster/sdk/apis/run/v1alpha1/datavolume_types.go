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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DataVolumeKind                = "DataVolume"
	DataVolumeIdLabel             = "run.ai/datavolume-id"
	DataVolumeNameLabel           = "run.ai/datavolume-name"
	DataVolumePvForNamespaceLabel = "run.ai/datavolume-pv-for-ns"
)

// DataVolumeSpec defines the desired state of DataVolume
type DataVolumeSpec struct {
	// the control-plane ID of the datavolume
	DatavolumeId string `json:"datavolumeId,omitempty"`

	// the description of the data volume
	Description string `json:"description,omitempty"`

	// the name of the PVC that the datavolume is based on
	OriginPvcName string `json:"originPvcName,omitempty"`

	// the name of the project that in its namespace the origin pvc is located
	ProjectName string `json:"projectName,omitempty"`

	// if true, the original storage volume will be deleted together with the datavolume
	ShouldDeleteOriginalVolume bool `json:"shouldDeleteOriginalVolume,omitempty"`

	// if true, the original storage volume will be used with the datavolume
	ShouldUseOriginalVolume bool `json:"shouldUseOriginalVolume,omitempty"`

	// the scopes that the datavolume is shared with
	SharedScopes []DataVolumeSharedScope `json:"sharedScopes,omitempty"`
}

// DataVolumeStatus defines the observed state of DataVolume
type DataVolumeStatus struct {
	// Phase is the recently observed phase of the data volume.
	Phase DataVolumePhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=ProjectPhase"`

	// A human-readable message indicating details about why the data volume is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// The name of the copied pvc in the origin namespace - in case the original is deleted
	DatavolumePvcName string `json:"datavolumePvcName,omitempty"`

	// The name of the copied pv that was attached to the original pvc - in case the original is deleted
	DatavolumePvName string `json:"datavolumePvName,omitempty"`

	// Current conditions of the data volume
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// DataVolume is the Schema for the datavolumes API
type DataVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataVolumeSpec   `json:"spec,omitempty"`
	Status DataVolumeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataVolumeList contains a list of DataVolume
type DataVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataVolume `json:"items"`
}

// +enum
type DataVolumePhase string

const (
	DataVolumeInitializing DataVolumePhase = "Initializing"
	DataVolumeReady        DataVolumePhase = "Ready"
	DataVolumeNotReady     DataVolumePhase = "NotReady"
	DataVolumeDeleting     DataVolumePhase = "Deleting"
)

type DataVolumeSharedScope struct {
	ScopeType ScopeType `json:"scopeType,omitempty"`
	ScopeName string    `json:"scopeName,omitempty"`
}

func init() {
	SchemeBuilder.Register(&DataVolume{}, &DataVolumeList{})
}

// IsOwner returns the index of this DataVolume in the given object's OwnerReferences array if it exists.
func (dataVolume *DataVolume) IsOwner(obj client.Object) (indexOfOwnerInRefArray int) {
	indexOfOwnerInRefArray = -1
	for index, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.Kind == DataVolumeKind &&
			ownerRef.UID == dataVolume.UID &&
			ownerRef.Name == dataVolume.Name {
			indexOfOwnerInRefArray = index
		}
	}
	return indexOfOwnerInRefArray
}

func (dataVolume *DataVolume) SetOwner(obj client.Object) {
	indexOfOwnerInRefArray := dataVolume.IsOwner(obj)
	if indexOfOwnerInRefArray == -1 {
		obj.SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: dataVolume.APIVersion,
			Kind:       dataVolume.Kind,
			Name:       dataVolume.Name,
			UID:        dataVolume.UID,
		}})
	}
}
