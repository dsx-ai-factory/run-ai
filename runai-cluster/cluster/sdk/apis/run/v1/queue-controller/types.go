/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package queue_controller

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

// QueueController defines configurations for the queue-controller
type QueueController struct {
	// Enabled specifies if the queue-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the queue-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Service describes the service for the queue-controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Resources describes the resource requirements for the pod grouper pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the queue-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the queue-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (r *QueueController) SetDefaultsWhereNeeded() {
	if r.Enabled == nil {
		r.Enabled = ptr.To(true)
	}

	if r.Resources == nil {
		r.Resources = &Resources{}
	}
	r.Resources.SetDefaultsWhereNeeded()

	if r.Service == nil {
		r.Service = &Service{}
	}
	r.Service.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("75m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("150Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("30m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("75Mi")
	}
}

// Service defines configuration for the queue-controller service
type Service struct {
	// Metrics specifies the metrics service spec
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`

	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook *Webhook `json:"webhook,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	if s.Metrics == nil {
		s.Metrics = &Metrics{}
	}
	s.Metrics.SetDefaultsWhereNeeded()

	if s.Webhook == nil {
		s.Webhook = &Webhook{}
	}
	s.Webhook.SetDefaultsWhereNeeded()
}

// Metrics defines configuration for the queue-controller metrics endpoint
type Metrics struct {
	// Port specifies the metrics service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the metrics service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`

	// Name specifies the name of the port
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`
}

func (p *Metrics) SetDefaultsWhereNeeded() {
	if p.Port == nil {
		p.Port = ptr.To(8080)
	}

	if p.TargetPort == nil {
		p.TargetPort = ptr.To(8080)
	}

	if p.Name == nil {
		p.Name = ptr.To("metrics")
	}
}

// Webhook defines configuration for the queue-controller webhook
type Webhook struct {
	// Port specifies the webhook service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the webhook service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

func (p *Webhook) SetDefaultsWhereNeeded() {
	if p.Port == nil {
		p.Port = ptr.To(443)
	}

	if p.TargetPort == nil {
		p.TargetPort = ptr.To(9443)
	}
}
