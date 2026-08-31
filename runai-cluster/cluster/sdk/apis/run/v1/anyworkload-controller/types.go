/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package anyworkload_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// AnyworkloadController defines configurations for the anyworkload-controller
type AnyworkloadController struct {
	// Enabled specifies if the anyworkload-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the anyworkload-controller image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the anyworkload-controller pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// UseDebugLogLevel states whether to use debug log level in the agent's logs
	// +kubebuilder:validation:Optional
	UseDebugLogLevel *bool `json:"useDebugLogLevel,omitempty"`

	// GPUNetworkAccelerationEnabled indicates if GPU network acceleration feature is enabled for the
	// anyworkload-controller.
	// the cluster must include the MNNVL prerequisites before enabling this flag
	// +kubebuilder:validation:Optional
	GPUNetworkAccelerationEnabled *bool `json:"GPUNetworkAccelerationEnabled,omitempty"`

	// EnableElasticComputeDomain enables elastic compute domain for MNNVL
	// +kubebuilder:validation:Optional
	EnableElasticComputeDomain *bool `json:"enableElasticComputeDomain,omitempty"`

	// StoragePluginEnabled indicates whether the anyworkload-controller storage plugin is enabled.
	// When enabled, the controller reconciles PVCs declared under spec.instructions.storage.newPvcs.
	// +kubebuilder:validation:Optional
	StoragePluginEnabled *bool `json:"storagePluginEnabled,omitempty"`

	// MaxMutationRounds specifies the maximum number of mutation rounds for resource mutation
	// +kubebuilder:validation:Optional
	MaxMutationRounds *int `json:"maxMutationRounds,omitempty"`

	// Metrics specifies the configuration of the metrics for the anyworkload-controller
	// +kubebuilder:validation:Optional
	Metrics *common.Service `json:"metrics,omitempty"`

	// RunaiOauth2ProxyImage specifies the image of oauth2-proxy, which is used to authenticate users for external URL workloads.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyImage *common.Image `json:"runaiOauth2ProxyImage,omitempty"`

	// RunaiOauth2ProxyRedisFQDN specifies the FQDN of Redis that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyRedisFQDN *string `json:"runaiOauth2ProxyRedisFQDN,omitempty"`

	// RunaiOauth2ProxyWorkloadAuthSecretName specifies the secret name that is used by the oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyWorkloadAuthSecretName *string `json:"runaiOauth2ProxyWorkloadAuthSecretName,omitempty"`

	// RunaiOauth2ProxySessionStoreType specifies the session store type that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxySessionStoreType *string `json:"runaiOauth2ProxySessionStoreType,omitempty"`

	// RunaiOauth2ProxyUsernameLoaderImage specifies the image of the username loader side container that is used by oauth2-proxy.
	// +kubebuilder:validation:Optional
	RunaiOauth2ProxyUsernameLoaderImage *common.Image `json:"runaiOauth2ProxyUsernameLoaderImage,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the anyworkload-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (a *AnyworkloadController) SetDefaultsWhereNeeded(globalRegistry string) {
	if a.Enabled == nil {
		a.Enabled = ptr.To(true)
	}

	if a.Image == nil {
		a.Image = &common.Image{}
	}
	a.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("anyworkload-controller"),
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

	if a.GPUNetworkAccelerationEnabled == nil {
		a.GPUNetworkAccelerationEnabled = ptr.To(false)
	}

	if a.EnableElasticComputeDomain == nil {
		a.EnableElasticComputeDomain = ptr.To(false)
	}

	if a.StoragePluginEnabled == nil {
		a.StoragePluginEnabled = ptr.To(true)
	}

	if a.MaxMutationRounds == nil {
		a.MaxMutationRounds = ptr.To(5)
	}

	if a.Metrics == nil {
		a.Metrics = &common.Service{}
	}
	a.Metrics.SetDefaultsWhereNeeded(9400, 9400, v1.ServiceTypeClusterIP)

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
