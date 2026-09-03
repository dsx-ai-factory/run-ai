/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package shared_objects_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// SharedObjectsController defines configurations for the shared-objects-controller
type SharedObjectsController struct {
	// Enabled specifies if the shared-objects-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the shared-objects-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the shared-objects-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// BaseControllerDelayMilliseconds specifies the base delay in milliseconds for the controller
	// +kubebuilder:validation:Optional
	BaseControllerDelayMilliseconds *int `json:"baseControllerDelayMilliseconds,omitempty"`

	// MaxControllerDelaySeconds specifies the maximum delay in seconds for the controller
	// +kubebuilder:validation:Optional
	MaxControllerDelaySeconds *int `json:"maxControllerDelaySeconds,omitempty"`

	// Replicas specifies the number of replicas of the shared objects controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the shared-objects-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the shared-objects-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *SharedObjectsController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("shared-objects-controller"),
	})
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

	if a.BaseControllerDelayMilliseconds == nil {
		a.BaseControllerDelayMilliseconds = ptr.To(500)
	}

	if a.MaxControllerDelaySeconds == nil {
		a.MaxControllerDelaySeconds = ptr.To(60)
	}

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("150m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("1024Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("50m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
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
