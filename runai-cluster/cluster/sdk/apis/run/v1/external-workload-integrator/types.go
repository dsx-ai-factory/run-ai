/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package external_workload_integrator

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

const (
	EnableRbacEscalationDefault = true
)

// ExternalWorkloadIntegrator defines configurations for the external-workload-integrator
type ExternalWorkloadIntegrator struct {
	// Enabled specifies if the external-workload-integrator should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the external-workload-integrator image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the external-workload-integrator pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// Replicas specifies the number of replicas of the external workload integrator
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Metrics defines the configuration of the external-workload-integrator's metrics endpoint
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`

	RIDRbacConfig *RIDRBACConfig `json:"ridRbacConfig,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the external-workload-integrator
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the external-workload-integrator
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`

	// Network specifies network configurations
	// +kubebuilder:validation:Optional
	Network *Network `json:"network,omitempty"`
}

// Metrics defines configuration of the external-workload-integrator's metrics endpoint
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
func (a *ExternalWorkloadIntegrator) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To[bool](true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("external-workload-integrator"),
	})

	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()

	if a.Metrics == nil {
		a.Metrics = &Metrics{}
	}
	a.Metrics.SetDefaultsWhereNeeded()

	if a.K8sClientConfig == nil {
		a.K8sClientConfig = &K8sClientConfig{}
	}
	a.K8sClientConfig.SetDefaultsWhereNeeded()

	if a.UseDebugLogLevel == nil {
		a.UseDebugLogLevel = ptr.To[bool](false)
	}

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}

	if a.RIDRbacConfig == nil {
		a.RIDRbacConfig = &RIDRBACConfig{}
	}
	a.RIDRbacConfig.SetDefaultsWhereNeeded()

	if a.Network == nil {
		a.Network = &Network{}
	}
	a.Network.SetDefaultsWhereNeeded()
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
		r.Limits[v1.ResourceMemory] = resource.MustParse("1024Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("100m")
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
		c.QPS = ptr.To[int](50)
	}

	if c.Burst == nil {
		c.Burst = ptr.To[int](300)
	}
}

type RIDRBACConfig struct {
	// Enabled specifies if the RID RBAC handler is enabled
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// ClusterRoleReadName specifies the name of the read cluster role that will be handled for the RID RBAC handler
	// +kubebuilder:validation:Optional
	ClusterRoleReadName string `json:"clusterRoleReadName,omitempty"`

	// ClusterRoleCrudName specifies the name of the CRUD cluster role that will be handled for the RID RBAC handler
	// +kubebuilder:validation:Optional
	ClusterRoleCrudName string `json:"clusterRoleCrudName,omitempty"`
}

func (r *RIDRBACConfig) SetDefaultsWhereNeeded() {
	if r.Enabled == nil {
		r.Enabled = ptr.To[bool](EnableRbacEscalationDefault)
	}

	if r.ClusterRoleReadName == "" {
		r.ClusterRoleReadName = "runai-additional-workload-gvks-read"
	}
	if r.ClusterRoleCrudName == "" {
		r.ClusterRoleCrudName = "runai-additional-workload-gvks-crud"
	}
}

type Network struct {
	// EndpointsLbTlsExplicitAnnotation specifies whether loadBalance service endpoint tls will be checked by looking at service annotations
	// +kubebuilder:validation:Optional
	LbServiceEndpointsTlsByExplicitAnnotation *bool `json:"lbServiceEndpointsTlsByExplicitAnnotation,omitempty"`
}

func (n *Network) SetDefaultsWhereNeeded() {
	if n.LbServiceEndpointsTlsByExplicitAnnotation == nil {
		n.LbServiceEndpointsTlsByExplicitAnnotation = ptr.To[bool](false)
	}
}
