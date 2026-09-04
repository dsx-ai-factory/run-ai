/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package admission_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// AdmissionController is the configuration struct for the Run:AI admission controller
type AdmissionController struct {
	// Enabled specifies if the admission-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the admission controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Replicas specifies the number of replicas of the admission controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Server describes the configurations of the admission controller server
	// +kubebuilder:validation:Optional
	Server *Server `json:"server,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Resources describes the resource requirements for the admission controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the admission-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *AdmissionController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To[bool](true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("admission-controller"),
	})
	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}

	if a.Server == nil {
		a.Server = &Server{}
	}
	a.Server.SetDefaultsWhereNeeded()

	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()
}

// Server specifies configuration for the admission controller.
type Server struct {
	// ReadTimeout is the timeout in seconds for the admission controller to read HTTP requests
	// +kubebuilder:validation:Optional
	ReadTimeout *int `json:"readTimeout,omitempty"`

	// WriteTimeout is the timeout in seconds for the admission controller to write HTTP responses
	// +kubebuilder:validation:Optional
	WriteTimeout *int `json:"writeTimeout,omitempty"`
}

func (s *Server) SetDefaultsWhereNeeded() {
	if s.ReadTimeout == nil {
		s.ReadTimeout = ptr.To(10)
	}
	if s.WriteTimeout == nil {
		s.WriteTimeout = ptr.To(10)
	}
}

// Args is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: these Core-era arguments are not consumed in KAI-based releases.
type Args struct {
	// +kubebuilder:validation:Optional
	RunAISharedDirectoryHostPath string `json:"runaiSharedDirectoryHostPath,omitempty"`
	// +kubebuilder:validation:Optional
	RunAIFractionalMinAllocationEnforcementBytes int `json:"runaiFractionalMinAllocationEnforcementBytes,omitempty"`
	// +kubebuilder:validation:Optional
	RunAIMpsDirectoryHostPath string `json:"runaiMpsDirectoryHostPath,omitempty"`
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("1000m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("500m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}
