/*
Copyright 2021 The KServe Authors.

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

// Package v1beta1 is a minimal local stub of a subset of the KServe
// serving v1beta1 API. The upstream module (github.com/kserve/kserve) pins
// itself back to an older k8s version via a replace directive and therefore
// cannot build against the k8s 0.35 line used by this repository. Only the
// fields/types exercised by the external-workload-integrator tests are
// reproduced here.
//
// +kubebuilder:object:generate=true
// +groupName=serving.kserve.io
package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "serving.kserve.io", Version: "v1beta1"}

	// SchemeGroupVersion is group version used to register these objects.
	SchemeGroupVersion = GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&InferenceService{}, &InferenceServiceList{})
}

// InferenceService is the Schema for the InferenceServices API.
type InferenceService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceServiceSpec   `json:"spec,omitempty"`
	Status InferenceServiceStatus `json:"status,omitempty"`
}

// InferenceServiceList contains a list of InferenceService.
type InferenceServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceService `json:"items"`
}

// InferenceServiceSpec is the top level type for this resource.
type InferenceServiceSpec struct {
	Predictor PredictorSpec `json:"predictor"`
}

// PredictorSpec defines the configuration for a predictor.
type PredictorSpec struct {
	// Model spec for any arbitrary framework.
	Model *ModelSpec `json:"model,omitempty"`

	// PodSpec fields shared by all predictors (e.g. Affinity, NodeSelector).
	corev1.PodSpec `json:",inline"`

	ComponentExtensionSpec `json:",inline"`
}

// ComponentExtensionSpec defines the deployment configuration for a given InferenceService component.
type ComponentExtensionSpec struct {
	// Annotations that will be added to the component pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ModelSpec describes a model to be served.
type ModelSpec struct {
	// ModelFormat being served.
	ModelFormat ModelFormat `json:"modelFormat"`

	PredictorExtensionSpec `json:",inline"`
}

// ModelFormat describes a model format.
type ModelFormat struct {
	// Name of the model format.
	Name string `json:"name"`
	// Version of the model format.
	// +optional
	Version *string `json:"version,omitempty"`
}

// PredictorExtensionSpec defines configuration shared across all predictor frameworks.
type PredictorExtensionSpec struct {
	// StorageURI points to the location of the trained model.
	// +optional
	StorageURI *string `json:"storageUri,omitempty"`
	// Container enables overrides for the predictor.
	corev1.Container `json:",inline"`
}

// InferenceServiceStatus defines the observed state of InferenceService.
type InferenceServiceStatus struct {
	duckv1.Status `json:",inline"`
}
