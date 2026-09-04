/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package init_ca

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// InitCA defines the configuration of the init-ca component
type InitCA struct {
	// Enabled specifies if init-ca should be run
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the init-ca image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the init-ca pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Replicas specifies the number of replicas of the init-ca service
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the init-ca
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (i *InitCA) SetDefaultsWhereNeeded(internalSpec *internal.Internal, globalReplicaCount *int32) {
	if i.Enabled == nil {
		if internalSpec.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOpenShift {
			i.Enabled = ptr.To[bool](false)
		} else {
			i.Enabled = ptr.To[bool](true)
		}
	}

	if i.Image == nil {
		i.Image = &common.Image{}
	}
	i.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(internalSpec.Global.ImageRegistry),
		Name:     ptr.To("init-ca-base"),
	})
	if i.Replicas == nil {
		if globalReplicaCount != nil {
			i.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			i.Replicas = ptr.To[int32](1)
		}
	}
	if i.Resources == nil {
		i.Resources = &Resources{}
	}
	i.Resources.SetDefaultsWhereNeeded()
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
