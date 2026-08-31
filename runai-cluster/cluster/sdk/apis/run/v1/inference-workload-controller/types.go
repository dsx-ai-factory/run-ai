/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package inference_workload_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// InferenceWorkloadController defines the configurations for the inference-workload-controller
type InferenceWorkloadController struct {
	// Enabled specifies if the inference-workload-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the inference-workload-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the inference-workload-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// Replicas specifies the number of replicas of the workload controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// GitSyncImage specifies the image of gitSync for the workload controller
	// +kubebuilder:validation:Optional
	GitSyncImage *common.Image `json:"gitSyncImage,omitempty"`

	// RunaiOauth2ProxyImage specifies the image of oauth2-proxy, which is used to authenticate users for external URL workloads.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyImage *common.Image `json:"runaiOauth2ProxyImage,omitempty"`

	// RunaiOauth2ProxyRedisImage specifies the image of Redis that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyRedisImage *common.Image `json:"runaiOauth2ProxyRedisImage,omitempty"`

	// RunaiOauth2ProxyRedisFQDN specifies the FQDN of Redis that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyRedisFQDN *string `json:"runaiOauth2ProxyRedisFQDN,omitempty"`

	// RunaiOauth2ProxyWorkloadAuthSecretName specifies the secret name that is used by the oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyWorkloadAuthSecretName *string `json:"runaiOauth2ProxyWorkloadAuthSecretName,omitempty"`

	// RunaiOauth2ProxyUsernameLoaderImage specifies the image of the username loader side container that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyUsernameLoaderImage *common.Image `json:"runaiOauth2ProxyUsernameLoaderImage,omitempty"`

	// RunaiOauth2ProxySessionStoreType specifies the session store type that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxySessionStoreType *string `json:"runaiOauth2ProxySessionStoreType,omitempty"`

	// RunaiLlmGatewayImage specifies the image of litellm, which is used in workloads that require a llm-gateway container.
	// +kubebuilder:validation:Optional
	RunaiLlmGatewayImage *common.Image `json:"runaiLlmGatewayImage,omitempty"`

	// RunaiLlmGatewayResourcesCpuCore specifies the CPU core resources of the llm-gateway container of a workload, if such.
	// +kubebuilder:validation:Optional
	RunaiLlmGatewayResourcesCpuCore *string `json:"runaiLlmGatewayResourcesCpuCore,omitempty"`

	// RunaiLlmGatewayResourcesCpuMemory specifies the CPU memory resources of the llm-gateway container of a workload, if such.
	// +kubebuilder:validation:Optional
	RunaiLlmGatewayResourcesCpuMemory *string `json:"runaiLlmGatewayResourcesCpuMemory,omitempty"`

	// ExternalAuthUrlEnabled indicates whether to use OAuth2 proxy when submitting workloads with an external URL.
	// default is true.
	// +kubebuilder:validation:Optional
	ExternalAuthUrlEnabled *bool `json:"externalAuthUrlEnabled,omitempty"`

	// ApplyControlPlanePolicies specifies whether or not to apply control plane policies to workloads
	// +kubebuilder:validation:Optional
	ApplyControlPlanePolicies *bool `json:"applyControlPlanePolicies,omitempty"`

	// GPUNetworkAccelerationEnabled indicates if GPU network acceleration feature is enabled for the inference-workload-controller.
	// the cluster must include the MNNVL prerequisites before enabling this flag
	// +kubebuilder:validation:Optional
	GPUNetworkAccelerationEnabled *bool `json:"GPUNetworkAccelerationEnabled,omitempty"`

	// Webhook specifies the service configuration for the webhook
	// +kubebuilder:validation:Optional
	Webhook *common.Service `json:"service,omitempty"`
	// HealthProbe specifies the service configuration for the health probe
	// +kubebuilder:validation:Optional
	HealthProbe *common.Service `json:"health,omitempty"`
	// Metric specifies the service configuration for the metric
	// +kubebuilder:validation:Optional
	Metric *common.Service `json:"metrics,omitempty"`

	// k8sClientBurst specifies the burst rate for the k8s client used by k8s
	// +kubebuilder:validation:Optional
	K8sClientBurst *string `json:"k8sClientBurst,omitempty"`

	// k8sClientQPS specifies the QPS rate for the k8s client used by k8s
	// +kubebuilder:validation:Optional
	K8sClientQPS *string `json:"k8sClientQPS,omitempty"`

	// EnableTopology indicates if network topology optimization feature is enabled for the inference-workload-controller.
	// +kubebuilder:validation:Optional
	EnableTopology *bool `json:"enableTopology,omitempty"`

	// EnableElasticComputeDomain indicates if elastic compute domain feature is enabled for the inference-workload-controller.
	// +kubebuilder:validation:Optional
	EnableElasticComputeDomain *bool `json:"enableElasticComputeDomain,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the inference-workload-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the inference-workload-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *InferenceWorkloadController) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if a.Enabled == nil {
		// for now, we always disable the inference-workload-controller
		a.Enabled = ptr.To(true)
	}

	if a.GPUNetworkAccelerationEnabled == nil {
		a.GPUNetworkAccelerationEnabled = ptr.To(false)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}

	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("inference-workload-controller"),
	})

	if a.Replicas == nil {
		if globalReplicaCount != nil {
			a.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			a.Replicas = ptr.To[int32](1)
		}
	}
	if a.Resources == nil {
		a.Resources = &Resources{}
	}
	a.Resources.SetDefaultsWhereNeeded()

	if a.K8sClientConfig == nil {
		a.K8sClientConfig = &K8sClientConfig{}
	}
	a.K8sClientConfig.SetDefaultsWhereNeeded()

	if a.GitSyncImage == nil {
		a.GitSyncImage = &common.Image{}
	}

	a.GitSyncImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(common.GitSyncImageRegistry),
		Name:     ptr.To(common.GitSyncImageName),
		Tag:      ptr.To(common.GitSyncImageTag),
	})

	if a.RunaiOauth2ProxyImage == nil {
		a.RunaiOauth2ProxyImage = &common.Image{}
	}
	a.RunaiOauth2ProxyImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To(common.WorkloadRunaiOauth2ProxyImageName),
	})

	if a.RunaiOauth2ProxyRedisFQDN == nil {
		a.RunaiOauth2ProxyRedisFQDN = ptr.To("cluster-redis.runai.svc.cluster.local:6379")
	}

	if a.RunaiOauth2ProxyWorkloadAuthSecretName == nil {
		a.RunaiOauth2ProxyWorkloadAuthSecretName = ptr.To(common.WorkloadRunaiOauth2ProxyWorkloadAuthSecretName)
	}

	if a.RunaiOauth2ProxySessionStoreType == nil {
		a.RunaiOauth2ProxySessionStoreType = ptr.To(common.WorkloadRunaiOauth2ProxyDefaultSessionStoreType)
	}

	if a.RunaiOauth2ProxyUsernameLoaderImage == nil {
		a.RunaiOauth2ProxyUsernameLoaderImage = &common.Image{}
	}
	a.RunaiOauth2ProxyUsernameLoaderImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageRegistry),
		Name:     ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageName),
		Tag:      ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageTag),
	})

	if a.RunaiLlmGatewayImage == nil {
		a.RunaiLlmGatewayImage = &common.Image{}
	}
	a.RunaiLlmGatewayImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To(common.WorkloadRunaiLlmGatewayImageName),
	})

	if a.RunaiLlmGatewayResourcesCpuCore == nil {
		a.RunaiLlmGatewayResourcesCpuCore = ptr.To(common.WorkloadRunaiLlmGatewayResourcesCpuCore)
	}

	if a.RunaiLlmGatewayResourcesCpuMemory == nil {
		a.RunaiLlmGatewayResourcesCpuMemory = ptr.To(common.WorkloadRunaiLlmGatewayResourcesCpuMemory)
	}

	if a.UseDebugLogLevel == nil {
		a.UseDebugLogLevel = ptr.To(false)
	}
	if a.Webhook == nil {
		a.Webhook = &common.Service{}
	}
	a.Webhook.SetDefaultsWhereNeeded(443, 9443, v1.ServiceTypeClusterIP)

	if a.HealthProbe == nil {
		a.HealthProbe = &common.Service{}
	}
	a.HealthProbe.SetDefaultsWhereNeeded(8081, 8081, v1.ServiceTypeClusterIP)
	if a.Metric == nil {
		a.Metric = &common.Service{}
	}
	a.Metric.SetDefaultsWhereNeeded(8383, 8383, v1.ServiceTypeClusterIP)

	if a.EnableTopology == nil {
		a.EnableTopology = ptr.To(true)
	}

	if a.EnableElasticComputeDomain == nil {
		a.EnableElasticComputeDomain = ptr.To(false)
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
