/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package nodepool_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// NodePoolController defines configuration for the Run:AI nodepool-controller component
type NodePoolController struct {
	// Enabled specifies if nodepool-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the nodepool-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the nodepool-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Args specify CLI arguments
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Service describes the service for the nodepool controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Replicas specifies the number of replicas of the node pool controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Metrics defines the configuration of the node pool controller's metrics endpoint
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the nodepool-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the nodepool-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// Metrics defines configuration of the node pool controller's metrics endpoint
type Metrics struct {
	// Port defines the port number for serving metrics
	// +kubebuilder:validation:Optional
	Port int32 `json:"port,omitempty"`

	// PortName defines the name of the port that serves metrics
	// +kubebuilder:validation:Optional
	PortName string `json:"portName,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (m *Metrics) SetDefaultsWhereNeeded() {
	if m.Port == 0 {
		m.Port = 9400
	}

	if m.PortName == "" {
		m.PortName = "metrics"
	}
}

// SetDefaultsWhereNeeded sets default for unset fields
func (c *NodePoolController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if c.Enabled == nil {
		c.Enabled = ptr.To[bool](true)
	}

	if c.Image == nil {
		c.Image = &common.Image{}
	}
	c.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("nodepool-controller"),
	})

	if c.Replicas == nil {
		if globalReplicaCount != nil {
			c.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			c.Replicas = ptr.To[int32](1)
		}
	}

	if c.Resources == nil {
		c.Resources = &Resources{}
	}
	c.Resources.SetDefaultsWhereNeeded()

	if c.Args == nil {
		c.Args = &Args{}
	}
	c.Args.SetDefaultsWhereNeeded()

	if c.Service == nil {
		c.Service = &Service{}
	}
	c.Service.SetDefaultsWhereNeeded()

	if c.Metrics == nil {
		c.Metrics = &Metrics{}
	}
	c.Metrics.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("900m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("1Gi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("450m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("512Mi")
	}
}

// Args specifies CLI arguments
type Args struct {
	// DebugLog specifies if debug messages should be printed
	// +kubebuilder:validation:Optional
	DebugLog bool `json:"debugLog,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {}

// Service defines configuration for the nodepool-controller service
type Service struct {
	// Type specifies the service type
	// +kubebuilder:validation:Optional
	Type v1.ServiceType `json:"type,omitempty"`

	// Port specifies the service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the target port in the container
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

func (service *Service) SetDefaultsWhereNeeded() {
	if service.Type == "" {
		service.Type = v1.ServiceTypeClusterIP
	}
	if service.Port == nil {
		service.Port = ptr.To[int](8080)
	}
	if service.TargetPort == nil {
		service.TargetPort = ptr.To[int](8080)
	}
}
