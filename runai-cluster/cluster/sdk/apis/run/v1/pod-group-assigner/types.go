/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package pod_group_assigner

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// PodGroupAssigner defines configurations for the pod-group-assigner
type PodGroupAssigner struct {
	// Enabled specifies if the pod-group-assigner should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the project controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the pod-group-assigner
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Service describes the service for the pod-group-assigner
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Resources describes the resource requirements for the pod-group-assigner pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// Replicas specifies the number of replicas of the pod group assigner
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the pod-group-assigner
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the pod-group-assigner
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *PodGroupAssigner) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	// KRM owns the pod-group-assigner now, so this operand only exists to delete
	// what the Run:ai-built one left behind. Overwritten rather than defaulted:
	// re-enabling it would admit every PodGroup twice, through two webhooks with
	// independent failure policies.
	p.Enabled = ptr.To[bool](false)

	if p.Image == nil {
		p.Image = &common.Image{}
	}
	p.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("pod-group-assigner"),
	})

	if p.Replicas == nil {
		if globalReplicaCount != nil {
			p.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			p.Replicas = ptr.To[int32](1)
		}
	}

	if p.Args == nil {
		p.Args = &Args{}
	}
	p.Args.SetDefaultsWhereNeeded()

	if p.Service == nil {
		p.Service = &Service{}
	}
	p.Service.SetDefaultsWhereNeeded()

	if p.Resources == nil {
		p.Resources = &Resources{}
	}
	p.Resources.SetDefaultsWhereNeeded()

	if p.K8sClientConfig == nil {
		p.K8sClientConfig = &K8sClientConfig{}
	}
	p.K8sClientConfig.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("200m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("100m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}

// Args defines command line arguments for the pod-grouper
type Args struct {
	// DebugLog specifies if debug messages should be printed
	// +kubebuilder:validation:Optional
	DebugLog bool `json:"debugLog,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {
}

// Service defines configuration for the pod-grouper service
type Service struct {
	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook Webhook `json:"webhook,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	s.Webhook.SetDefaultsWhereNeeded()
}

// Webhook defines configuration for the pod-grouper webhook
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

// K8sClientConfig specifies the configuration of k8s client
type K8sClientConfig struct {
	// Burst specifies the burst rate for the k8s client
	// +kubebuilder:validation:Optional
	Burst *int `json:"burst,omitempty"`

	// QPS specifies the QPS rate for the k8s client
	// +kubebuilder:validation:Optional
	QPS *int `json:"qps,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (c *K8sClientConfig) SetDefaultsWhereNeeded() {
	if c.QPS == nil {
		c.QPS = ptr.To[int](50)
	}

	if c.Burst == nil {
		c.Burst = ptr.To[int](300)
	}
}
