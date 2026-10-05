/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package inference_endpoint_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

type InferenceEndpointController struct {
	// Enabled controls installation of the controller. Gateway API and GIE CRDs must already exist.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

type Metrics struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`
	// +kubebuilder:validation:Optional
	PortName string `json:"portName,omitempty"`
}

type Resources v1.ResourceRequirements

func (c *InferenceEndpointController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if c.Enabled == nil {
		// TODO(viper): Revisit enabled-by-default before merging; Gateway API and GIE are prerequisites.
		c.Enabled = ptr.To(true)
	}
	if c.Image == nil {
		c.Image = &common.Image{}
	}
	c.Image.SetDefaultsWhereNeeded(&common.Image{Registry: ptr.To(globalRegistry), Name: ptr.To("inference-endpoint-controller")})
	if c.Resources == nil {
		c.Resources = &Resources{}
	}
	if c.Resources.Requests == nil {
		c.Resources.Requests = v1.ResourceList{}
	}
	if c.Resources.Limits == nil {
		c.Resources.Limits = v1.ResourceList{}
	}
	for key, value := range (v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m"), v1.ResourceMemory: resource.MustParse("128Mi")}) {
		if _, exists := c.Resources.Requests[key]; !exists {
			if limit, exists := c.Resources.Limits[key]; exists && value.Cmp(limit) > 0 {
				value = limit.DeepCopy()
			}
			c.Resources.Requests[key] = value
		}
	}
	for key, value := range (v1.ResourceList{v1.ResourceCPU: resource.MustParse("500m"), v1.ResourceMemory: resource.MustParse("256Mi")}) {
		if _, exists := c.Resources.Limits[key]; !exists {
			if request := c.Resources.Requests[key]; request.Cmp(value) > 0 {
				value = request.DeepCopy()
			}
			c.Resources.Limits[key] = value
		}
	}
	if c.Replicas == nil {
		c.Replicas = ptr.To(ptr.Deref(globalReplicaCount, 1))
	}
	if c.UseDebugLogLevel == nil {
		c.UseDebugLogLevel = ptr.To(false)
	}
	if c.Metrics == nil {
		c.Metrics = &Metrics{}
	}
	if c.Metrics.Port == 0 {
		c.Metrics.Port = 8080
	}
	if c.Metrics.PortName == "" {
		c.Metrics.PortName = "metrics"
	}
}
