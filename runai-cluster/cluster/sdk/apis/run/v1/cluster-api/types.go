/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package cluster_api

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// ClusterApi defines configurations for the cluster-api
type ClusterApi struct {
	// Enabled specifies if the cluster-api should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the cluster-api image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the cluster-api pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// NodePortService specifies the node port of the cluster-api service
	// +kubebuilder:validation:Optional
	NodePortService *Service `json:"service,omitempty"`

	// Ingress specifies the ingress configuration of the cluster-api
	// +kubebuilder:validation:Optional
	Ingress *common.Ingress `json:"ingress,omitempty"`

	// GatewayAPI specifies a per-operand Gateway reference override.
	// When set, this takes precedence over the global gatewayAPI name/namespace.
	// +kubebuilder:validation:Optional
	GatewayAPI *common.GatewayRef `json:"gatewayAPI,omitempty"`

	// Route specifies the route configuration of the cluster-api
	// This is only relevant for OpenShift
	// +kubebuilder:validation:Optional
	Route *common.Route `json:"route,omitempty"`
	// Replicas specifies the number of replicas for the cluster-api
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// WriteTimeout specifies the time in seconds for router write timeout
	// +kubebuilder:validation:Optional
	WriteTimeout *int32 `json:"writeTimeout,omitempty"`
	// ReadTimeout specifies the time in seconds for router read timeout
	// +kubebuilder:validation:Optional
	ReadTimeout *int32 `json:"readTimeout,omitempty"`
	// IdleTimeout specifies the time in seconds for router idle timeout
	// +kubebuilder:validation:Optional
	IdleTimeout *int32 `json:"idleTimeout,omitempty"`
	// KeepaliveIdleTimeSec specifies the time in seconds for keepalive idle check
	// +kubebuilder:validation:Optional
	KeepaliveIdleTimeSec *int32 `json:"keepaliveIdleTimeSec,omitempty"`
	// KeepalivePingTimeSec specifies the time in seconds for keepalive ping check
	// +kubebuilder:validation:Optional
	KeepalivePingTimeSec *int32 `json:"keepalivePingTimeSec,omitempty"`
	// CacheExpirationPeriodSec specifies how often cached auth and authz values should expire from the cache
	// +kubebuilder:validation:Optional
	CacheExpirationPeriodSec *int32 `json:"cacheExpirationPeriodSec,omitempty"`
	// WorkloadAuthCacheTtlSec specifies the TTL in seconds for the workload-authorization
	// cache used by the k8s-object endpoints (per (subject, workload) pair).
	// +kubebuilder:validation:Optional
	WorkloadAuthCacheTtlSec *int32 `json:"workloadAuthCacheTtlSec,omitempty"`
	// Metric specifies the service configuration for the metric
	// +kubebuilder:validation:Optional
	Metrics *common.Service `json:"metrics,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the cluster-api
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the cluster-api
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *ClusterApi) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(true)
	}

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("cluster-api"),
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
	if a.NodePortService == nil {
		a.NodePortService = &Service{}
	}
	a.NodePortService.SetDefaultsWhereNeeded()
	if a.Ingress == nil {
		a.Ingress = &common.Ingress{}
	}
	a.Ingress.SetDefaultsWhereNeeded()
	if a.Route == nil {
		a.Route = &common.Route{}
	}
	a.Route.SetDefaultsWhereNeeded()
	if a.WriteTimeout == nil {
		a.WriteTimeout = ptr.To[int32](15)
	}
	if a.ReadTimeout == nil {
		a.ReadTimeout = ptr.To[int32](15)
	}
	if a.IdleTimeout == nil {
		a.IdleTimeout = ptr.To[int32](60)
	}
	if a.KeepaliveIdleTimeSec == nil {
		a.KeepaliveIdleTimeSec = ptr.To[int32](300)
	}
	if a.KeepalivePingTimeSec == nil {
		a.KeepalivePingTimeSec = ptr.To[int32](5)
	}
	if a.CacheExpirationPeriodSec == nil {
		a.CacheExpirationPeriodSec = ptr.To[int32](180)
	}
	if a.WorkloadAuthCacheTtlSec == nil {
		a.WorkloadAuthCacheTtlSec = ptr.To[int32](60)
	}
	if a.Metrics == nil {
		a.Metrics = &common.Service{}
	}
	a.Metrics.SetDefaultsWhereNeeded(9400, 9400, v1.ServiceTypeClusterIP)
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
		r.Limits[v1.ResourceMemory] = resource.MustParse("1024Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("500m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("512Mi")
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

type Service struct {
	// NodePort specifies the node port of the researcher-service service
	// +kubebuilder:validation:Optional
	NodePort int `json:"nodePort,omitempty"`
}

func (s *Service) SetDefaultsWhereNeeded() {
	if s.NodePort == 0 {
		s.NodePort = 32283
	}
}
