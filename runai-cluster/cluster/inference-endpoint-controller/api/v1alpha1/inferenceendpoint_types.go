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
	ReasonPending          = "Pending"
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
	// endpoints on the same Gateway never collide on the shared /v1 path.
	// +required
	// +kubebuilder:validation:MinItems=1
	// +listType=set
	Hostnames []string `json:"hostnames"`
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

// SchedulingProfileName identifies a preloaded scheduling profile.
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=63
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
type SchedulingProfileName string

// SchedulingProfile defines one preloaded combination of scoring weights.
type SchedulingProfile struct {
	// name is referenced by defaultProfile and activeProfile.
	// +required
	Name SchedulingProfileName `json:"name"`

	// scorers replace the scoring selection for this profile. At least one
	// weight must be positive. Catalog changes roll the EPP Deployment.
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:XValidation:rule="self.exists(s, s.weight > 0)",message="at least one scorer weight must be positive"
	// +listType=map
	// +listMapKey=type
	Scorers []ScorerSpec `json:"scorers"`
}

// SchedulingSpec tunes pod selection. Static weights and the default profile
// are loaded at EPP startup. Only activeProfile changes apply without a rollout.
// +kubebuilder:validation:XValidation:rule="!has(self.profiles) || !has(self.scorers)",message="profiles and scorers cannot be combined"
// +kubebuilder:validation:XValidation:rule="!has(self.profiles) || (has(self.defaultProfile) && has(self.activeProfile))",message="profiles require defaultProfile and activeProfile"
// +kubebuilder:validation:XValidation:rule="!has(self.defaultProfile) || (has(self.profiles) && self.profiles.exists(p, p.name == self.defaultProfile))",message="defaultProfile must name a configured profile"
// +kubebuilder:validation:XValidation:rule="!has(self.activeProfile) || (has(self.profiles) && self.profiles.exists(p, p.name == self.activeProfile))",message="activeProfile must name a configured profile"
type SchedulingSpec struct {
	// scorers and their weights. Defaults to queue 30, kvCache 20, prefixCache 10.
	// +optional
	// +listType=map
	// +listMapKey=type
	Scorers []ScorerSpec `json:"scorers,omitempty"`

	// profiles is the preloaded catalog for llm-d-v0.10; it replaces scorers.
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	// +listType=map
	// +listMapKey=name
	Profiles []SchedulingProfile `json:"profiles,omitempty"`

	// defaultProfile is the EPP fallback for a missing or blank profile header.
	// Changing it updates the ConfigMap and rolls the EPP Deployment.
	// +optional
	DefaultProfile SchedulingProfileName `json:"defaultProfile,omitempty"`

	// activeProfile is set by the Gateway in the EPP-Profile request header.
	// Switching to another preloaded profile updates only the HTTPRoute.
	// +optional
	ActiveProfile SchedulingProfileName `json:"activeProfile,omitempty"`
}

// EppProfile identifies a supported endpoint-picker configuration contract.
// +kubebuilder:validation:Enum=gie-v1.5;llm-d-v0.10
type EppProfile string

const (
	// EppProfileGIEV15 selects the GIE v1.5.0 stack covered by the Track A lab.
	EppProfileGIEV15 EppProfile = "gie-v1.5"
	// EppImageGIEV15 is the image supported by EppProfileGIEV15.
	// Keep the EppSpec image-compatibility validation in sync with this value.
	EppImageGIEV15 = "registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0"
	// EppProfileLLMDV010 opts into the experimental upstream header profile handler.
	EppProfileLLMDV010 EppProfile = "llm-d-v0.10"
	// EppImageLLMDV010 is the digest verified by manual E2E case 04.
	// Keep the EppSpec image-compatibility validation in sync with this value.
	EppImageLLMDV010 = "ghcr.io/llm-d/llm-d-router-endpoint-picker:v0.10.0@sha256:2e516fa1310da7be59b82beb1445362139597d6d553ef04d546716abe3aaaa70"
)

// EppSpec configures the endpoint picker the controller deploys.
// +kubebuilder:validation:XValidation:rule="has(self.profile) || (has(self.image) && self.image == 'registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0')",message="image-only EPP requires the approved GIE v1.5.0 image; otherwise select a supported profile"
// +kubebuilder:validation:XValidation:rule="has(self.profile) || has(self.image)",message="epp.profile or epp.image is required"
// +kubebuilder:validation:XValidation:rule="!has(self.profile) || !has(self.image) || (self.profile == 'gie-v1.5' && self.image == 'registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0') || (self.profile == 'llm-d-v0.10' && self.image == 'ghcr.io/llm-d/llm-d-router-endpoint-picker:v0.10.0@sha256:2e516fa1310da7be59b82beb1445362139597d6d553ef04d546716abe3aaaa70')",message="epp.image must match the selected profile's pinned image"
// +kubebuilder:validation:XValidation:rule="!(has(self.profile) && self.profile == 'llm-d-v0.10') || (has(self.scheduling) && has(self.scheduling.profiles))",message="llm-d-v0.10 requires scheduling.profiles"
// +kubebuilder:validation:XValidation:rule="!has(self.scheduling) || !has(self.scheduling.profiles) || (has(self.profile) && self.profile == 'llm-d-v0.10')",message="scheduling.profiles requires the llm-d-v0.10 profile"
type EppSpec struct {
	// profile selects a versioned configuration, image, startup flags and ports.
	// gie-v1.5 reproduces Track A. llm-d-v0.10 supports experimental selection
	// among preloaded scheduling profiles with a Gateway-controlled header.
	// +optional
	Profile EppProfile `json:"profile,omitempty"`

	// image overrides the profile image only when it matches the supported image.
	// Without a profile, only EppImageGIEV15 is allowed; it receives the
	// GIE v1.5 configuration and startup flags.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image,omitempty"`

	// scheduling tunes pod selection.
	// +optional
	Scheduling *SchedulingSpec `json:"scheduling,omitempty"`
}

// RewriteTarget is one destination of a model rewrite rule.
type RewriteTarget struct {
	// model name sent to the pod in place of the requested one.
	// +required
	// +kubebuilder:validation:MinLength=1
	Model string `json:"model"`

	// weight is the share of requests for this target, relative to the other
	// targets of the same rule. Weights do not need to sum to 100.
	// +required
	// +kubebuilder:validation:Minimum=1
	Weight int32 `json:"weight"`
}

// RewriteSpec maps one requested model name to one or more served model
// names. It becomes a rule in an InferenceModelRewrite the EPP watches at
// runtime, so changes apply without restarting anything.
type RewriteSpec struct {
	// model is the exact name clients put in the request body.
	// +required
	// +kubebuilder:validation:MinLength=1
	Model string `json:"model"`

	// targets and their traffic shares. One target is an alias, several are a canary.
	// +required
	// +kubebuilder:validation:MinItems=1
	Targets []RewriteTarget `json:"targets"`
}

// ObjectiveSpec is one priority lane. It becomes an InferenceObjective the EPP
// reads at runtime: when the pool is saturated, requests in a higher-priority
// lane are served first and lower lanes are queued or shed.
type ObjectiveSpec struct {
	// model labels the lane and gives the generated InferenceObjective its name:
	// <endpoint>-objective-<model as a DNS label>. The EPP does not read the
	// model name from the request body for this; a client selects the lane by
	// sending that generated name in the x-gateway-inference-objective request
	// header. Requests without the header get priority 0.
	// +required
	// +kubebuilder:validation:MinLength=1
	Model string `json:"model"`

	// priority of this lane relative to the other lanes of the same pool. 0 is
	// the value the EPP assumes for requests without a lane, so lanes cannot
	// go below it.
	// +required
	// +kubebuilder:validation:Minimum=0
	Priority int32 `json:"priority"`
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

	// rewrites are model aliases and canary splits, applied live by the EPP.
	// +optional
	// +listType=map
	// +listMapKey=model
	Rewrites []RewriteSpec `json:"rewrites,omitempty"`

	// objectives are priority lanes per model name, applied live by the EPP.
	// +optional
	// +listType=map
	// +listMapKey=model
	Objectives []ObjectiveSpec `json:"objectives,omitempty"`
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
