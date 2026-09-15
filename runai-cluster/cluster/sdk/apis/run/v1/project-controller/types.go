/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package project_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// ProjectController defines configurations for the project-controller
type ProjectController struct {
	// Enabled specifies if the project-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the project controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the project controller
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Service describes the service for the project controller
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// IsOpenShift indicates that this cluster is managed by OpenShift
	// +kubebuilder:validation:Optional
	IsOpenShift bool `json:"isOpenShift,omitempty"`

	// CreateNamespaces specifies if a namespace should be created for
	// a new project, if false, users have to create a namespace themselves
	// and link it to the project
	// +kubebuilder:validation:Optional
	CreateNamespaces *bool `json:"createNamespaces,omitempty"`

	// CreateRoleBindings specifies if role bindings should be created
	// in the project's namespace
	// +kubebuilder:validation:Optional
	CreateRoleBindings *bool `json:"createRoleBindings,omitempty"`

	// CreateServiceMonitors specifies if service monitors should be created
	// in the project's namespace
	// +kubebuilder:validation:Optional
	CreateServiceMonitors *bool `json:"createServiceMonitors,omitempty"`

	// ClusterWideSecret specifies if cluster-wide secrets should be copied
	// from the runai namespace to a project's namespace
	// +kubebuilder:validation:Optional
	ClusterWideSecret *bool `json:"clusterWideSecret,omitempty"`

	// ClusterWideConfigMap specifies if cluster-wide ConfigMap should be copied
	// from the runai namespace to a project's namespace
	// +kubebuilder:validation:Optional
	ClusterWideConfigMap *bool `json:"clusterWideConfigMap,omitempty"`

	// ClusterWidePvc specifies if cluster-wide pvcs should be copied
	// from the runai namespace to a project's namespace
	// +kubebuilder:validation:Optional
	ClusterWidePvc *bool `json:"clusterWidePvc,omitempty"`

	// LimitRange specifies if limit ranges should be defined for projects
	// +kubebuilder:validation:Optional
	LimitRange *bool `json:"limitRange,omitempty"`

	// Resources describes the resource requirements for the project controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Profiling specifies profiling configuration
	// +kubebuilder:validation:Optional
	Profiling *Profiling `json:"profiling,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// Replicas specifies the number of replicas of the project controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the project-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *ProjectController) SetDefaultsWhereNeeded(internalSpec *internal.Internal, globalRegistry string, globalReplicaCount *int32) {
	// KRM owns the project-controller now, so this operand only exists to delete
	// what the Run:ai-built one left behind. Overwritten rather than defaulted:
	// re-enabling it would run a second controller against the same Projects.
	p.Enabled = ptr.To[bool](false)

	p.IsOpenShift = internalSpec.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOpenShift

	if p.Image == nil {
		p.Image = &common.Image{}
	}
	p.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("project-controller"),
	})
	if p.Args == nil {
		p.Args = &Args{}
	}
	p.Args.SetDefaultsWhereNeeded()

	if p.Service == nil {
		p.Service = &Service{}
	}
	p.Service.SetDefaultsWhereNeeded()

	if p.CreateNamespaces == nil {
		p.CreateNamespaces = ptr.To[bool](true)
	}

	if p.CreateRoleBindings == nil {
		p.CreateRoleBindings = ptr.To[bool](true)
	}

	if p.CreateServiceMonitors == nil {
		p.CreateServiceMonitors = ptr.To[bool](true)
	}

	if p.ClusterWideSecret == nil {
		p.ClusterWideSecret = ptr.To[bool](true)
	}

	if p.ClusterWideConfigMap == nil {
		p.ClusterWideConfigMap = ptr.To[bool](true)
	}

	if p.ClusterWidePvc == nil {
		p.ClusterWidePvc = ptr.To[bool](true)
	}

	if p.LimitRange == nil {
		p.LimitRange = ptr.To[bool](false)
	}

	if p.Resources == nil {
		p.Resources = &Resources{}
	}
	p.Resources.SetDefaultsWhereNeeded()

	if p.Profiling == nil {
		p.Profiling = &Profiling{}
	}
	p.Profiling.SetDefaultsWhereNeeded()

	if p.K8sClientConfig == nil {
		p.K8sClientConfig = &K8sClientConfig{}
	}
	p.K8sClientConfig.SetDefaultsWhereNeeded()

	if p.Replicas == nil {
		if globalReplicaCount != nil {
			p.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			p.Replicas = ptr.To[int32](1)
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

// Args defines command line arguments for the project-controller
type Args struct {
	// Verbosity specifies the logging verbosity of the project controller
	// use 0 for info, -1 for debug
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (a *Args) SetDefaultsWhereNeeded() {
	if a.Verbosity == nil {
		a.Verbosity = ptr.To[int](0)
	}
}

// Service defines configuration for the project-controller service
type Service struct {
	// Type specifies the kubernetes ServiceType of this service
	// +kubebuilder:validation:Optional
	Type v1.ServiceType `json:"type,omitempty"`

	// Metrics specifies the metrics service spec
	// +kubebuilder:validation:Optional
	Metrics Metrics `json:"metrics,omitempty"`

	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook Webhook `json:"webhook,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (s *Service) SetDefaultsWhereNeeded() {
	if s.Type == "" {
		s.Type = v1.ServiceTypeClusterIP
	}

	s.Metrics.SetDefaultsWhereNeeded()

	s.Webhook.SetDefaultsWhereNeeded()
}

// Metrics defines configuration for the project-controller metrics endpoint
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

// Webhook defines configuration for the project-controller webhook
type Webhook struct {
	// Port specifies the webhook service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the webhook service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (w *Webhook) SetDefaultsWhereNeeded() {
	if w.Port == nil {
		w.Port = ptr.To[int](9443)
	}

	if w.TargetPort == nil {
		w.TargetPort = ptr.To[int](9443)
	}
}

type Profiling struct {
	// Enabled specifies whether profiling is enabled
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	ApiPort string `json:"apiPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default fields for unset fields
func (p *Profiling) SetDefaultsWhereNeeded() {
	if p.Enabled == nil {
		p.Enabled = ptr.To[bool](false)
	}

	if p.ApiPort == "" {
		p.ApiPort = "8182"
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
