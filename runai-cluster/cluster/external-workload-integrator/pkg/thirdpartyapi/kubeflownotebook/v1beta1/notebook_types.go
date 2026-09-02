/*
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

// Package v1beta1 is a minimal local stub of the Kubeflow notebook-controller
// kubeflow.org/v1beta1 Notebook API. The upstream module
// (github.com/kubeflow/kubeflow/components/notebook-controller) ships a webhook
// using the controller-runtime v0.22 single-argument ctrl.NewWebhookManagedBy,
// which does not compile against controller-runtime v0.23 (generic, two-argument
// builder). Upstream has no release compatible with cr v0.23 / k8s 0.35, and the
// run:ai code only consumes the Notebook type (not the webhook), so the few
// types/helpers we need are reproduced here.
//
// +kubebuilder:object:generate=true
// +groupName=kubeflow.org
package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "kubeflow.org", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Notebook{}, &NotebookList{})
}

// NotebookSpec defines the desired state of Notebook.
type NotebookSpec struct {
	// Template describes the notebooks that will be created.
	Template NotebookTemplateSpec `json:"template,omitempty"`
}

type NotebookTemplateSpec struct {
	Spec corev1.PodSpec `json:"spec,omitempty"`
}

// NotebookStatus defines the observed state of Notebook.
type NotebookStatus struct {
	// Conditions is an array of current conditions.
	Conditions []NotebookCondition `json:"conditions"`
	// ReadyReplicas is the number of Pods created by the StatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas"`
	// ContainerState is the state of underlying container.
	ContainerState corev1.ContainerState `json:"containerState"`
}

type NotebookCondition struct {
	// Type is the type of the condition. Possible values are Running|Waiting|Terminated.
	Type string `json:"type"`
	// Status is the status of the condition. Can be True, False, Unknown.
	Status string `json:"status"`
	// LastProbeTime is the last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// LastTransitionTime is the last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason is the (brief) reason the container is in the current state.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message regarding why the container is in the current state.
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Notebook is the Schema for the notebooks API.
type Notebook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotebookSpec   `json:"spec,omitempty"`
	Status NotebookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotebookList contains a list of Notebook.
type NotebookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notebook `json:"items"`
}
