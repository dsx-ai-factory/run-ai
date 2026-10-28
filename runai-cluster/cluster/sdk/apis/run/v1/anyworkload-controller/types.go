/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package anyworkload_controller

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// DefaultS3FileSystemImageName identifies the bundled S3 filesystem image.
const DefaultS3FileSystemImageName = "goofys"

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

	// GitSyncImage specifies the image used by the git-sync injected container.
	// +kubebuilder:validation:Optional
	GitSyncImage *common.Image `json:"gitSyncImage,omitempty"`

	// S3FileSystemImage specifies the image used by the S3 filesystem injected container.
	// +kubebuilder:validation:Optional
	S3FileSystemImage *common.Image `json:"s3FileSystemImage,omitempty"`

	// S3FileSystemMountCheckEnabled overrides startup-probe support for the configured S3 image.
	// When unset, only the bundled goofys image from this operator build is enabled.
	// Custom images must implement goofys --ready-file and --check-ready before opting in.
	// Readiness markers use a writable volume mounted only into the S3 sidecar.
	// +kubebuilder:validation:Optional
	S3FileSystemMountCheckEnabled *bool `json:"s3FileSystemMountCheckEnabled,omitempty"`

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

	// ExternalAuthUrlEnabled indicates whether external URL workloads use the OAuth2 proxy.
	// +kubebuilder:validation:Optional
	ExternalAuthUrlEnabled *bool `json:"externalAuthUrlEnabled,omitempty"`

	// OpenshiftRouteConfiguration controls defaults for Run:ai-managed OpenShift Route objects.
	// +kubebuilder:validation:Optional
	OpenshiftRouteConfiguration *OpenshiftRouteConfiguration `json:"openshiftRouteConfiguration,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the anyworkload-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// OpenshiftRouteTerminationPolicy is the OpenShift Route TLS termination policy applied to Run:ai-managed Routes.
// +kubebuilder:validation:Enum=edge;passthrough
type OpenshiftRouteTerminationPolicy string

const (
	OpenshiftRouteTerminationPolicyEdge        OpenshiftRouteTerminationPolicy = "edge"
	OpenshiftRouteTerminationPolicyPassthrough OpenshiftRouteTerminationPolicy = "passthrough"
)

// OpenshiftRouteConfiguration configures defaults applied to Run:ai-managed OpenShift Route objects.
type OpenshiftRouteConfiguration struct {
	// TerminationPolicy specifies the TLS termination policy for Run:ai-generated OpenShift Routes.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=edge
	TerminationPolicy OpenshiftRouteTerminationPolicy `json:"terminationPolicy,omitempty" jsonschema:"enum=edge,enum=passthrough,default=edge"`
}

func (o *OpenshiftRouteConfiguration) GetTerminationPolicy() OpenshiftRouteTerminationPolicy {
	if o == nil || o.TerminationPolicy == "" {
		return OpenshiftRouteTerminationPolicyEdge
	}
	return o.TerminationPolicy
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

	if a.OpenshiftRouteConfiguration == nil {
		a.OpenshiftRouteConfiguration = &OpenshiftRouteConfiguration{}
	}
	if a.OpenshiftRouteConfiguration.TerminationPolicy == "" {
		a.OpenshiftRouteConfiguration.TerminationPolicy = OpenshiftRouteTerminationPolicyEdge
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

	if a.GitSyncImage == nil {
		a.GitSyncImage = &common.Image{}
	}
	a.GitSyncImage.SetDefaultsWhereNeeded(common.ThirdPartyImage("GIT_SYNC"))

	if a.S3FileSystemImage == nil {
		a.S3FileSystemImage = &common.Image{}
	}
	a.S3FileSystemImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To(DefaultS3FileSystemImageName),
	})

	if a.RunaiOauth2ProxyImage == nil {
		a.RunaiOauth2ProxyImage = &common.Image{}
	}
	a.RunaiOauth2ProxyImage.SetDefaultsWhereNeeded(common.ThirdPartyImage("OAUTH2_PROXY"))

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
	a.RunaiOauth2ProxyUsernameLoaderImage.SetDefaultsWhereNeeded(common.ThirdPartyImage("WAIT_UBI_MINIMAL"))

	if a.ExternalAuthUrlEnabled == nil {
		a.ExternalAuthUrlEnabled = ptr.To(true)
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
