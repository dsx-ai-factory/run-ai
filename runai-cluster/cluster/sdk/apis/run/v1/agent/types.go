/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package agent

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

// Agent defines the configuration of the Run:AI agent
type Agent struct {
	// Enabled specifies if the agent should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// SyncIntervalMillis is the interval in which
	// the agent syncs data to the backend
	// +kubebuilder:validation:Optional
	SyncIntervalMillis *int `json:"syncIntervalMillis,omitempty"`

	// AccessRulesLimit is the pagination limit to use when
	// the agent gets AccessRules from the backend
	// +kubebuilder:validation:Optional
	AccessRulesLimit *int `json:"accessRulesLimit,omitempty"`

	// Image is the configuration of the agent image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the agent pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Metrics defines the configuration of the agent's metrics endpoint
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`

	// RejectUnauthorized decides if the backend should be accessed with TLS
	// +kubebuilder:validation:Optional
	RejectUnauthorized *bool `json:"rejectUnauthorized,omitempty"`

	// APIVersion is the version of the RunAI API
	// +kubebuilder:validation:Optional
	APIVersion string `json:"apiVersion,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	EnableWorkloadRunner *bool `json:"enableWorkloadRunner,omitempty"`

	// Replicas specifies the number of replicas of the agent service
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the agent
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the agent
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *Agent) SetDefaultsWhereNeeded(internalSpec *runv1_internal.Internal, globalReplicaCount *int32) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(!internalSpec.RemoteConfig.GetSystemConfig().Deactivated)
	}

	if a.SyncIntervalMillis == nil {
		a.SyncIntervalMillis = ptr.To[int](common.AgentSyncIntervalMillisDefault)
	}

	if a.AccessRulesLimit == nil {
		a.AccessRulesLimit = ptr.To[int](400)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(internalSpec.Global.ImageRegistry),
		Name:     ptr.To("agent"),
	})
	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()

	if a.Metrics == nil {
		a.Metrics = &Metrics{}
	}
	a.Metrics.SetDefaultsWhereNeeded()

	if a.RejectUnauthorized == nil {
		a.RejectUnauthorized = ptr.To[bool](true)
	}

	if a.APIVersion == "" {
		a.APIVersion = "v1"
	}

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
		r.Limits[v1.ResourceCPU] = resource.MustParse("700m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("1Gi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("350m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("512Mi")
	}
}

// Metrics defines configuration of the agent's metrics endpoint
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
