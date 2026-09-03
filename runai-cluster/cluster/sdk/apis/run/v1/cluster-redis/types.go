/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package cluster_redis

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

const (
	redisPort = 6379
)

// ClusterRedis defines the configuration of the cluster redis
type ClusterRedis struct {
	// Enabled specifies if the cluster redis should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image specifies the redis image to use
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Replicas specifies the number of replicas of the cluster redis
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resources specifies the resource requirements for the cluster redis
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// RedisService specifies the service configuration
	// +kubebuilder:validation:Optional
	RedisService *common.Service `json:"redisService,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the cluster-redis
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the cluster-redis
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *ClusterRedis) SetDefaultsWhereNeeded(internalSpec *internal.Internal, global *global.Global) {
	if r.Enabled == nil {
		r.Enabled = ptr.To[bool](true)
	}

	if r.Image == nil {
		r.Image = &common.Image{}
	}
	r.Image.SetDefaultsWhereNeeded(common.ThirdPartyImage("VALKEY"))
	if r.Replicas == nil {
		if global.ReplicaCount != nil {
			r.Replicas = ptr.To[int32](*global.ReplicaCount)
		} else {
			r.Replicas = ptr.To[int32](1)
		}
	}
	if r.Resources == nil {
		r.Resources = &Resources{}
	}
	r.Resources.SetDefaultsWhereNeeded()

	if r.RedisService == nil {
		r.RedisService = &common.Service{}
	}
	r.RedisService.Port = ptr.To(redisPort)
	r.RedisService.TargetPort = ptr.To(redisPort)
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

func (r *Resources) SetDefaultsWhereNeeded() {
	if r.Limits == nil {
		r.Limits = v1.ResourceList{}
	}

	if r.Requests == nil {
		r.Requests = v1.ResourceList{}
	}

	if _, exists := r.Limits[v1.ResourceCPU]; !exists {
		r.Limits[v1.ResourceCPU] = resource.MustParse("600m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("300m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}
