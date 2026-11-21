/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Copyright 2026.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// Condition types reported in InferenceEndpointStatus.Conditions.
const (
	// ConditionReady is True when every child object exists, the HTTPRoute is
	// accepted by its Gateway, and the current EPP Deployment rollout is complete.
	ConditionReady = "Ready"
	// ConditionRouteAccepted mirrors the Accepted and ResolvedRefs conditions the
	// Gateway reports on the generated HTTPRoute.
	ConditionRouteAccepted = "RouteAccepted"
)

// Condition reasons used with the condition types above.
const (
	ReasonRouteAccepted    = "RouteAccepted"
	ReasonRouteNotAccepted = "RouteNotAccepted"
	ReasonApplyFailed      = "ApplyFailed"
	ReasonEppNotAvailable  = "EppNotAvailable"
)

// WorkloadSelector picks pods by label, in the same shape the InferencePool uses.
type WorkloadSelector struct {
	// matchLabels must all be present on a pod for it to be selected.
	// +required
	// +kubebuilder:validation:MinProperties=1
	MatchLabels map[string]string `json:"matchLabels"`
}

// WorkloadSpec identifies the model-serving pods. The label selector is the only
// contract between the endpoint and the pods; the controller never creates or
// modifies pods.
type WorkloadSpec struct {
	// selector picks the pods that serve this endpoint. It becomes the
	// InferencePool selector.
	// +required
	Selector WorkloadSelector `json:"selector"`

	// port is the container port the pods listen on. The EPP returns
	// podIP:port to the Gateway.
	// +required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// GatewayReference names an existing Gateway. The controller attaches an
// HTTPRoute to it and never creates or modifies the Gateway itself.
type GatewayReference struct {
	// name of the Gateway.
	// +required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// namespace of the Gateway. Defaults to the namespace of the
	// InferenceEndpoint.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// RoutingSpec describes how requests reach the endpoint.
type RoutingSpec struct {
	// gatewayRef is the Gateway the generated HTTPRoute attaches to.
	// +required
	GatewayRef GatewayReference `json:"gatewayRef"`

	// hostnames the HTTPRoute matches. At least one is required so that two
	// endpoints on the same Gateway never collide on the shared path prefix.
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:items:MaxLength=253
	// +kubebuilder:validation:items:Pattern=`^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	// +listType=set
	Hostnames []string `json:"hostnames"`

	// pathPrefix is the request path whose requests go through the EPP.
	// Requests outside it do not match the route and return 404 at the
	// Gateway.
	// +optional
	// +kubebuilder:default="/v1"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^/`
	PathPrefix string `json:"pathPrefix,omitempty"`
}

// ScorerType names one of the endpoint picker's scoring signals.
// +kubebuilder:validation:Enum=queue;kvCache;prefixCache
type ScorerType string

const (
	// ScorerQueue prefers pods with fewer queued requests (GIE queue-scorer).
	ScorerQueue ScorerType = "queue"
	// ScorerKVCache prefers pods with more free KV cache (GIE kv-cache-utilization-scorer).
	ScorerKVCache ScorerType = "kvCache"
	// ScorerPrefixCache prefers pods that already hold the prompt prefix (GIE prefix-cache-scorer).
	ScorerPrefixCache ScorerType = "prefixCache"
)

// ScorerSpec is one scoring signal and how much it counts.
type ScorerSpec struct {
	// type of the scorer.
	// +required
	Type ScorerType `json:"type"`

	// weight of this scorer relative to the others. 0 keeps the scorer loaded
	// but ignores its opinion.
	// +required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Weight int32 `json:"weight"`
}

// SchedulingSpec tunes pod selection.
type SchedulingSpec struct {
	// scorers and their weights. Defaults to queue 30, kvCache 20, prefixCache 10.
	// +optional
	// +listType=map
	// +listMapKey=type
	Scorers []ScorerSpec `json:"scorers,omitempty"`
}

// EppProfile identifies a supported endpoint-picker configuration contract.
// +kubebuilder:validation:Enum=gie-v1.5
type EppProfile string

const (
	// EppProfileGIEV15 selects the GIE v1.5.0 endpoint picker.
	EppProfileGIEV15 EppProfile = "gie-v1.5"
	// EppImageGIEV15 is the image pinned by EppProfileGIEV15.
	EppImageGIEV15 = "registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0"
)

// EppSpec configures the endpoint picker the controller deploys.
type EppSpec struct {
	// profile selects a versioned configuration, image, startup flags and ports.
	// +required
	Profile EppProfile `json:"profile"`

	// scheduling tunes pod selection.
	// +optional
	Scheduling *SchedulingSpec `json:"scheduling,omitempty"`
}

// InferenceEndpointSpec defines the desired state of InferenceEndpoint.
type InferenceEndpointSpec struct {
	// workload identifies the serving pods.
	// +required
	Workload WorkloadSpec `json:"workload"`

	// routing describes the Gateway attachment.
	// +required
	Routing RoutingSpec `json:"routing"`

	// epp configures the endpoint picker.
	// +required
	Epp EppSpec `json:"epp"`
}

// InferenceEndpointStatus defines the observed state of InferenceEndpoint.
type InferenceEndpointStatus struct {
	// conditions represent the current state of the InferenceEndpoint.
	// Types: Ready, RouteAccepted.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration is the spec generation the conditions describe.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Gateway",type=string,JSONPath=`.spec.routing.gatewayRef.name`
// +kubebuilder:printcolumn:name="Hostnames",type=string,JSONPath=`.spec.routing.hostnames`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// InferenceEndpoint is the Schema for the inferenceendpoints API.
// One InferenceEndpoint produces one routing stack: an InferencePool selecting
// the workload pods, an endpoint picker deployment, and an HTTPRoute on an
// existing Gateway.
type InferenceEndpoint struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of InferenceEndpoint
	// +required
	Spec InferenceEndpointSpec `json:"spec"`

	// status defines the observed state of InferenceEndpoint
	// +optional
	Status InferenceEndpointStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// InferenceEndpointList contains a list of InferenceEndpoint
type InferenceEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []InferenceEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &InferenceEndpoint{}, &InferenceEndpointList{})
		return nil
	})
}
