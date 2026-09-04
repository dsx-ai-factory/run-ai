/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package application_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// ApplicationController defines configurations for the application-controller
type ApplicationController struct {
	// Enabled specifies if the application-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the application-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Replicas specifies the number of replicas of the application controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Service describes the service for the application controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Resources describes the resource requirements for the application-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the application-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *ApplicationController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("application-controller"),
	})

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}

	if a.Service == nil {
		a.Service = &Service{}
	}
	a.Service.SetDefaultsWhereNeeded()

	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()

	if a.K8sClientConfig == nil {
		a.K8sClientConfig = &K8sClientConfig{}
	}
	a.K8sClientConfig.SetDefaultsWhereNeeded()

	if a.UseDebugLogLevel == nil {
		a.UseDebugLogLevel = ptr.To(false)
	}
}

// Service defines configuration for the application-controller service
type Service struct {
	// Type specifies the kubernetes ServiceType of this service
	// +kubebuilder:validation:Optional
	Type v1.ServiceType `json:"type,omitempty"`

	// Metrics specifies the metrics service spec
	// +kubebuilder:validation:Optional
	Metrics Metrics `json:"metrics,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	if s.Type == "" {
		s.Type = v1.ServiceTypeClusterIP
	}

	s.Metrics.SetDefaultsWhereNeeded()
}

// Metrics defines configuration for the application-controller metrics endpoint
type Metrics struct {
	// Port specifies the metrics service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// PortName defines the name of the port that serves metrics
	// +kubebuilder:validation:Optional
	PortName string `json:"portName,omitempty"`

	Test string `json:"test,omitempty"`
	// TargetPort specifies the metrics service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields for the project-controller metrics
func (m *Metrics) SetDefaultsWhereNeeded() {
	if m.Port == nil {
		m.Port = ptr.To(9400)
	}

	if m.PortName == "" {
		m.PortName = "metrics"
	}

	if m.TargetPort == nil {
		m.TargetPort = ptr.To(9400)
	}
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
		c.QPS = ptr.To(50)
	}

	if c.Burst == nil {
		c.Burst = ptr.To(300)
	}
}
