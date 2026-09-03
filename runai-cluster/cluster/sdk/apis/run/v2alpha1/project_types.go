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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:deprecatedversion

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

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	Department                  string             `json:"department,omitempty"`
	DeservedGpus                float64            `json:"deservedGpus,omitempty"`
	MaxAllowedGpus              float64            `json:"maxAllowedGpus,omitempty"`
	GpuOverQuotaWeight          float64            `json:"gpuOverQuotaWeight,omitempty"`
	InteractiveJobTimeLimitSecs int64              `json:"interactiveJobTimeLimitSecs,omitempty"`
	NodeAffinityInteractive     []string           `json:"nodeAffinityInteractive,omitempty"`
	NodeAffinityTrain           []string           `json:"nodeAffinityTrain,omitempty"`
	Permissions                 ProjectPermissions `json:"permissions,omitempty"`
	LimitsRef                   int64              `json:"limitsRef,omitempty"`
	SecretsRef                  map[string]string  `json:"secretsRef,omitempty"`
	CpuDefaultRequest           string             `json:"cpuDefaultRequest,omitempty"`
	MemoryDefaultRequest        string             `json:"memoryDefaultRequest,omitempty"`
	CpuDefaultLimit             string             `json:"cpuDefaultLimit,omitempty"`
	MemoryDefaultLimit          string             `json:"memoryDefaultLimit,omitempty"`
	CpuMaxLimit                 string             `json:"cpuMaxLimit,omitempty"`
	MemoryMaxLimit              string             `json:"memoryMaxLimit,omitempty"`
	SwapEnabled                 bool               `json:"swapEnabled,omitempty"`
	Resources                   *ProjectResources  `json:"resources,omitempty"`
}

type ProjectResources struct {
	GPU    GPUResources   `json:"gpu,omitempty"`
	CPU    SystemResource `json:"cpu,omitempty"`
	Memory SystemResource `json:"memory,omitempty"`
}

type GPUResources struct {
	// +optional
	Deserved float64 `json:"deserved"`
	// +optional
	OverQuotaWeight float64 `json:"overQuotaWeight"`
	// +optional
	Limit float64 `json:"limit"`
}

type SystemResource struct {
	// +optional
	// +kubebuilder:validation:XIntOrString
	Deserved int `json:"deserved"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type ProjectPermissions struct {
	Users  []string `json:"users,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}

// IsOwner returns the index of this Project in the given object's OwnerReferences array if it exists.
func (project Project) IsOwner(obj client.Object) (indexOfOwnerInRefArray int) {
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
