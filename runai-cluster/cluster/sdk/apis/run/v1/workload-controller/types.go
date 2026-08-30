// +kubebuilder:object:generate:=true
package workload_controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internalv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// WorkloadController defines the configuration of the workload-controller
type WorkloadController struct {
	// Enabled specifies if the workload-controller should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image specifies the image configuration for the workload controller
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Replicas specifies the number of replicas of the workload controller
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resources specifies the resource requirements for the workload controller
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// GitSyncImage specifies the image of gitSync for the workload controller
	// +kubebuilder:validation:Optional
	GitSyncImage *common.Image `json:"gitSyncImage,omitempty"`

	// S3FileSystemImage specifies the image of s3 file system for the workload controller
	// +kubebuilder:validation:Optional
	S3FileSystemImage *common.Image `json:"s3FileSystemImage,omitempty"`

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

	// k8sClientBurst specifies the burst rate for the k8s client used by k8s
	// +kubebuilder:validation:Optional
	K8sClientBurst *string `json:"k8sClientBurst,omitempty"`

	// k8sClientQPS specifies the QPS rate for the k8s client used by k8s
	// +kubebuilder:validation:Optional
	K8sClientQPS *string `json:"k8sClientQPS,omitempty"`

	// JupyterUserIdentity indicates whether to use jupyter user identity configuration when running jupyter workloads.
	// default is true. ref: https://jupyter-docker-stacks.readthedocs.io/en/latest/using/common.html#user-related-configurations.
	// +kubebuilder:validation:Optional
	JupyterUserIdentity *bool `json:"jupyterUserIdentity,omitempty"`

	// ExternalAuthUrlEnabled indicates whether to use OAuth2 proxy when submitting workloads with an external URL.
	// default is true.
	// +kubebuilder:validation:Optional
	ExternalAuthUrlEnabled *bool `json:"externalAuthUrlEnabled,omitempty"`

	// InteractiveMaxConcurrentReconciles specifies the maximum number of concurrent reconciles for interactive workloads
	// +kubebuilder:validation:Optional
	InteractiveMaxConcurrentReconciles *int `json:"interactiveMaxConcurrentReconciles,omitempty"`

	// TrainingMaxConcurrentReconciles specifies the maximum number of concurrent reconciles for training workloads
	// +kubebuilder:validation:Optional
	TrainingMaxConcurrentReconciles *int `json:"trainingMaxConcurrentReconciles,omitempty"`

	// DistributedMaxConcurrentReconciles specifies the maximum number of concurrent reconciles for distributed workloads
	// +kubebuilder:validation:Optional
	DistributedMaxConcurrentReconciles *int `json:"distributedMaxConcurrentReconciles,omitempty"`

	// DegradedPhaseDelay [Deprecated] specifies the phase delay in seconds from running to degraded
	// +kubebuilder:validation:Optional
	DegradedPhaseDelay *int `json:"degradedPhaseDelay,omitempty"`

	// GPUNetworkAccelerationEnabled indicates if GPU network acceleration feature is enabled for the workload-controller.
	// the cluster must include the MNNVL prerequisites before enabling this flag
	// +kubebuilder:validation:Optional
	GPUNetworkAccelerationEnabled *bool `json:"GPUNetworkAccelerationEnabled,omitempty"`

	// EnableTopology indicates if network topology optimization feature is enabled for the workload-controller.
	// +kubebuilder:validation:Optional
	EnableTopology *bool `json:"enableTopology,omitempty"`

	// EnableElasticComputeDomain indicates if elastic compute domain feature is enabled for the workload-controller.
	// +kubebuilder:validation:Optional
	EnableElasticComputeDomain *bool `json:"enableElasticComputeDomain,omitempty"`

	// EnableDanglingResourceCleanup indicates if automatic cleanup of dangling Kubernetes
	// resources left over from workload networking mode transitions is enabled.
	// +kubebuilder:validation:Optional
	EnableDanglingResourceCleanup *bool `json:"enableDanglingResourceCleanup,omitempty"`

	// OpenshiftRouteConfiguration controls defaults for Run:ai-managed OpenShift Route objects.
	// +kubebuilder:validation:Optional
	OpenshiftRouteConfiguration *OpenshiftRouteConfiguration `json:"openshiftRouteConfiguration,omitempty"`

	// Type specifies the service type
	ServicesType v1.ServiceType `json:"type,omitempty"`
	// WebhookAddress override the workload controller webhook access address in the cluster, used by the workload migration
	WebhookAddress *string `json:"webhookAddress,omitempty"`
	// Webhook specifies the service configuration for the webhook
	// +kubebuilder:validation:Optional
	Webhook *common.Service `json:"service,omitempty"`
	// HealthProbe specifies the service configuration for the health probe
	// +kubebuilder:validation:Optional
	HealthProbe *common.Service `json:"health,omitempty"`
	// Metric specifies the service configuration for the metric
	// +kubebuilder:validation:Optional
	Metric *common.Service `json:"metrics,omitempty"`
	// Profiler specifies the configuration for the profiler
	// +kubebuilder:validation:Optional
	Profiler *ProfilerConfig `json:"profiler,omitempty"`
	// SkipMigration specifies whether to skip the migration
	SkipMigration *bool `json:"skipMigration,omitempty"`

	// ApplyControlPlanePolicies specifies whether or not to apply control plane policies to workloads
	// +kubebuilder:validation:Optional
	ApplyControlPlanePolicies *bool `json:"applyControlPlanePolicies,omitempty"`

	// AdditionalJobLabels are labels that should be added to all job pods
	AdditionalPodLabels map[string]string `json:"additionalPodLabels,omitempty"`

	// AdditionalPodAnnotations are annotations that should be added to all job pods
	AdditionalPodAnnotations map[string]string `json:"additionalPodAnnotations,omitempty"`

	// FailureResourceCleanupPolicy is the policy for handling sub-resources cleanup in workload failure ('All' - default, 'None', 'KeepFailing')
	// This policy applies only to failures in workload infrastructure, not to cases where the workload gracefully completed with a failing result.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=All;None;KeepFailing
	FailureResourceCleanupPolicy *string `json:"failureResourceCleanupPolicy,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the workload-controller
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the workload-controller
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
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
	// "edge" terminates TLS at the OpenShift Router; "passthrough" forwards encrypted traffic to the pod for end-to-end TLS.
	// default is edge.
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

type ProfilerConfig struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	ServerAddress *string `json:"serverAddress,omitempty"`
}

func (p *ProfilerConfig) SetDefaultsWhereNeeded() {
	if p.Enabled == nil {
		p.Enabled = ptr.To(false)
	}
	if p.ServerAddress == nil {
		p.ServerAddress = ptr.To("localhost:6060")
	}
}

func (p *ProfilerConfig) IsEnabled() bool {
	if p == nil || p.Enabled == nil || *p.Enabled == false {
		return false
	}
	return true
}

func (p *ProfilerConfig) GetProfilerServerAddress() string {
	if p == nil || p.ServerAddress == nil {
		return ""
	}
	return *p.ServerAddress
}

type PodService struct {
	Disable *bool `json:"enabled,omitempty"`

	ServicePortConfig `json:",inline"`
}

func (p *PodService) IsEnabled() bool {
	if p == nil || p.Disable == nil || *p.Disable == false {
		return true
	}
	return false
}

type ServicePortConfig struct {

	// Port specifies the service port
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the target port in the container
	TargetPort *int `json:"targetPort,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (w *WorkloadController) SetDefaultsWhereNeeded(internalSpec *internalv1.Internal, globalSpec *global.Global) {
	globalRegistry := internalSpec.Global.ImageRegistry

	if w.Enabled == nil {
		w.Enabled = ptr.To[bool](true)
	}

	if w.Image == nil {
		w.Image = &common.Image{}
	}
	w.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("workload-controller"),
	})
	if w.Replicas == nil {
		if globalSpec.ReplicaCount != nil {
			w.Replicas = ptr.To[int32](*globalSpec.ReplicaCount)
		} else {
			w.Replicas = ptr.To[int32](1)
		}
	}
	if w.Resources == nil {
		w.Resources = &Resources{}
	}
	w.Resources.SetDefaultsWhereNeeded()

	if w.SkipMigration == nil {
		w.SkipMigration = ptr.To(false)
	}

	if w.GitSyncImage == nil {
		w.GitSyncImage = &common.Image{}
	}
	w.GitSyncImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(common.GitSyncImageRegistry),
		Name:     ptr.To(common.GitSyncImageName),
		Tag:      ptr.To(common.GitSyncImageTag),
	})

	if w.S3FileSystemImage == nil {
		w.S3FileSystemImage = &common.Image{}
	}
	w.S3FileSystemImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("goofys"),
	})

	if w.RunaiOauth2ProxyImage == nil {
		w.RunaiOauth2ProxyImage = &common.Image{}
	}
	w.RunaiOauth2ProxyImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To(common.WorkloadRunaiOauth2ProxyImageName),
	})

	if w.RunaiOauth2ProxyUsernameLoaderImage == nil {
		w.RunaiOauth2ProxyUsernameLoaderImage = &common.Image{}
	}
	w.RunaiOauth2ProxyUsernameLoaderImage.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageRegistry),
		Name:     ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageName),
		Tag:      ptr.To(common.WorkloadRunaiOauth2ProxyUsernameLoaderImageTag),
	})

	if w.JupyterUserIdentity == nil {
		w.JupyterUserIdentity = ptr.To(true)
	}
	if w.ExternalAuthUrlEnabled == nil {
		w.ExternalAuthUrlEnabled = ptr.To(true)
	}
	if w.ApplyControlPlanePolicies == nil {
		w.ApplyControlPlanePolicies = ptr.To(false)
	}
	if w.ServicesType == "" {
		w.ServicesType = v1.ServiceTypeClusterIP
	}
	if w.Webhook == nil {
		w.Webhook = &common.Service{}
	}
	w.Webhook.SetDefaultsWhereNeeded(443, 9443, v1.ServiceTypeClusterIP)

	if w.HealthProbe == nil {
		w.HealthProbe = &common.Service{}
	}
	w.HealthProbe.SetDefaultsWhereNeeded(8081, 8081, v1.ServiceTypeClusterIP)
	if w.Metric == nil {
		w.Metric = &common.Service{}
	}
	w.Metric.SetDefaultsWhereNeeded(8383, 8383, v1.ServiceTypeClusterIP)

	if w.Profiler == nil {
		w.Profiler = &ProfilerConfig{}
	}
	w.Profiler.SetDefaultsWhereNeeded()

	if w.RunaiOauth2ProxyRedisFQDN == nil {
		w.RunaiOauth2ProxyRedisFQDN = ptr.To("cluster-redis.runai.svc.cluster.local:6379")
	}

	if w.RunaiOauth2ProxyWorkloadAuthSecretName == nil {
		w.RunaiOauth2ProxyWorkloadAuthSecretName = ptr.To(common.WorkloadRunaiOauth2ProxyWorkloadAuthSecretName)
	}

	if w.RunaiOauth2ProxySessionStoreType == nil {
		w.RunaiOauth2ProxySessionStoreType = ptr.To(common.WorkloadRunaiOauth2ProxyDefaultSessionStoreType)
	}

	if w.GPUNetworkAccelerationEnabled == nil {
		w.GPUNetworkAccelerationEnabled = ptr.To(false)
	}

	if w.EnableTopology == nil {
		w.EnableTopology = ptr.To(true)
	}

	if w.EnableElasticComputeDomain == nil {
		w.EnableElasticComputeDomain = ptr.To(false)
	}

	if w.EnableDanglingResourceCleanup == nil {
		w.EnableDanglingResourceCleanup = ptr.To(true)
	}

	if w.OpenshiftRouteConfiguration == nil {
		w.OpenshiftRouteConfiguration = &OpenshiftRouteConfiguration{}
	}
	if w.OpenshiftRouteConfiguration.TerminationPolicy == "" {
		w.OpenshiftRouteConfiguration.TerminationPolicy = OpenshiftRouteTerminationPolicyEdge
	}
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
