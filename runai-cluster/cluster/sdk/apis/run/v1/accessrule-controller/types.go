/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package accessrule_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// AccessRuleController defines configurations for the accessrule-controller
type AccessRuleController struct {
	// Enabled specifies if the accessrule-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the accessrule-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the accessrule-controller
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Service describes the service for the accessrule-controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Resources describes the resource requirements for the accessrule-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Replicas specifies the number of replicas of the accessrule controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the accessrule-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *AccessRuleController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To[bool](true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}

	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("accessrule-controller"),
	})
	if a.Args == nil {
		a.Args = &Args{}
	}
	a.Args.SetDefaultsWhereNeeded()

	if a.Service == nil {
		a.Service = &Service{}
	}
	a.Service.SetDefaultsWhereNeeded()

	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

// SetDefaultsWhereNeeded sets default fields for unset fields
func (r *Resources) SetDefaultsWhereNeeded() {
	if r.Limits == nil {
		r.Limits = v1.ResourceList{}
	}

	if r.Requests == nil {
		r.Requests = v1.ResourceList{}
	}

	if _, exists := r.Limits[v1.ResourceCPU]; !exists {
		r.Limits[v1.ResourceCPU] = resource.MustParse("300m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("2Gi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("150m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("1Gi")
	}
}

// Args defines command line arguments for the accessrule-controller
type Args struct {
	// DebugLog specifies if debug messages should be printed
	// +kubebuilder:validation:Optional
	DebugLog bool `json:"debugLog,omitempty"`

	// LegacyProjectPermissions specifies if legacy project permissions should be handled
	// +kubebuilder:validation:Optional
	LegacyProjectPermissions *bool `json:"legacyProjectPermissions,omitempty"`

	// LegacyProjectPermissionsIntervalSeconds specifies the interval in seconds
	// for the handling of legacy project permissions
	LegacyProjectPermissionsIntervalSeconds *int `json:"legacyProjectPermissionsIntervalSeconds,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {
	if a.LegacyProjectPermissions == nil {
		// default to true
		a.LegacyProjectPermissions = ptr.To[bool](true)
	}

	if a.LegacyProjectPermissionsIntervalSeconds == nil {
		// default to 30 seconds
		a.LegacyProjectPermissionsIntervalSeconds = ptr.To[int](30)
	}
}

// Service defines configuration for the accessrule-controller
type Service struct {
	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook Webhook `json:"webhook,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	s.Webhook.SetDefaultsWhereNeeded()
}

// Webhook defines configuration for the accessrule-controller webhook
type Webhook struct {
	// Port specifies the webhook service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the webhook service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (w *Webhook) SetDefaultsWhereNeeded() {
	if w.Port == nil {
		w.Port = ptr.To[int](443)
	}

	if w.TargetPort == nil {
		w.TargetPort = ptr.To[int](8443)
	}
}
