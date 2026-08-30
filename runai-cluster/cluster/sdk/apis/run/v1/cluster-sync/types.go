// +kubebuilder:object:generate:=true
package cluster_sync

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

// CLusterSync defines configuration for the Run:AI cluster-sync component
type ClusterSync struct {
	// Enabled specifies if cluster-sync should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the cluster-sync image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the cluster-sync pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// SSLCertificateVerification indicates if SSL certificates should be verified
	// +kubebuilder:validation:Optional
	SSLCertificateVerification *bool `json:"SSLCertificateVerification,omitempty"`

	// Metrics defines the configuration of the cluster-sync's metrics endpoint
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`

	// MaxNumOfEventsPerRequest specifies the max amount of events to send in each request
	// +kubebuilder:validation:Optional
	MaxNumOfEventsPerRequest *int `json:"maxNumOfEventsPerRequest,omitempty"`

	// DEPRECATED MaxNumOfPodGroupsPerRequest specifies the max amount of pod groups to send in each request
	// +kubebuilder:validation:Optional
	MaxNumOfPodGroupsPerRequest *int `json:"maxNumOfPodGroupsPerRequest,omitempty"`

	// MaxNumOfPodsPerRequest specifies the max amount of pods to send in each request
	// +kubebuilder:validation:Optional
	MaxNumOfPodsPerRequest *int `json:"maxNumOfPodsPerRequest,omitempty"`

	// MaxNumOfWorkloadsPerRequest specifies the max amount of workloads to send in each request
	// +kubebuilder:validation:Optional
	MaxNumOfWorkloadsPerRequest *int `json:"maxNumOfWorkloadsPerRequest,omitempty"`

	// MaxNumOfDatavolumesPerRequest specifies the max amount of datavolumes to send in each request
	// +kubebuilder:validation:Optional
	MaxNumOfDatavolumesPerRequest *int `json:"maxNumOfDatavolumesPerRequest,omitempty"`

	// SyncEventsIntervalSeconds specifies the interval for syncing events to backend
	// performed by cluster-sync
	// +kubebuilder:validation:Optional
	SyncEventsIntervalSeconds *int `json:"syncEventsIntervalSeconds,omitempty"`

	// LivenessIntervalSeconds specifies the interval of liveness probes
	// performed by cluster-sync
	// +kubebuilder:validation:Optional
	LivenessIntervalSeconds *int `json:"livenessIntervalSeconds,omitempty"`

	// MaxNumberOfRetries specifies the maximum number of failed liveness
	// probes performed before the cluster-sync pod is restarted
	// +kubebuilder:validation:Optional
	MaxNumberOfRetries *int `json:"maxNumberOfRetries,omitempty"`

	// RetriesIntervalSeconds specifies the interval between failing
	// liveness probes
	// +kubebuilder:validation:Optional
	RetriesIntervalSeconds *int `json:"retriesIntervalSeconds,omitempty"`

	// SyncIntervalMinutes specifies the interval of periodic sync operation
	// +kubebuilder:validation:Optional
	SyncIntervalMinutes *int `json:"syncIntervalMinutes,omitempty"`

	// ExistingSecretToSyncLabel specifies the secret label that cluster-sync will use to find the secret
	// +kubebuilder:validation:Optional
	ExistingSecretToSyncLabel string `json:"existingSecretToSyncLabel,omitempty"`

	// ConfigMapToSyncLabel specifies the config map label that cluster-sync will use to find the config map
	// +kubebuilder:validation:Optional
	ConfigMapToSyncLabel string `json:"configMapToSyncLabel,omitempty"`

	// LogLevel specifies the log level of the cluster sync (-1=debug, 0=info, 1=warning, 2=error...)
	// +kubebuilder:validation:Optional
	LogLevel *int `json:"logLevel,omitempty"`

	// Profiling specifies profiling configuration
	// +kubebuilder:validation:Optional
	Profiling *Profiling `json:"profiling,omitempty"`

	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`

	// NvidiaGPULabel is the label to look for on GPU nodes (in case nfd is or is not installed)
	// +kubebuilder:validation:Optional
	NvidiaGPULabel *string `json:"nvidiaGPULabel,omitempty"`

	// NodeLabelNvidiaContainerToolkit is the label to look for on nodes that indicates container-toolkit is installed
	// +kubebuilder:validation:Optional
	NodeLabelNvidiaContainerToolkit *string `json:"nodeLabelNvidiaContainerToolkit,omitempty"`

	// NodeLabelNvidiaDcgmExporter is the label to look for on nodes that indicates nvidia-dcgm-exporter is installed
	// +kubebuilder:validation:Optional
	NodeLabelNvidiaDcgmExporter *string `json:"nodeLabelNvidiaDcgmExporter,omitempty"`

	// EnableNodeNvidiaComponentsReadiness - should enable readiness check for nvidia components on nodes
	// +kubebuilder:validation:Optional
	EnableNodeNvidiaComponentsReadiness *bool `json:"enableNodeNvidiaComponentsReadiness,omitempty"`

	// DEPRECATED UseNewWorkloadsApi specifies whether to use the new workloads api to sync workloads, pods and events
	// +kubebuilder:validation:Optional
	UseNewWorkloadsApi *bool `json:"useNewWorkloadsApi,omitempty"`

	// DEPRECATED UseDeprecatedWorkloadsApi specifies whether to use the old (deprecated) workloads api to sync workloads, pods and events
	// +kubebuilder:validation:Optional
	UseDeprecatedWorkloadsApi *bool `json:"useDeprecatedWorkloadsApi,omitempty"`

	// RateLimiterWorkloadsDelaySeconds specifies the rate limiter delay for workload-like scale objects
	// +kubebuilder:validation:Optional
	RateLimiterWorkloadsDelaySeconds *int `json:"rateLimiterWorkloadsDelaySeconds,omitempty"`

	// RateLimiterWorkloadsNumOfWorkers specifies the rate limiter number of workers for workload-like scale objects
	// +kubebuilder:validation:Optional
	RateLimiterWorkloadsNumOfWorkers *int `json:"rateLimiterWorkloadsNumOfWorkers,omitempty"`

	// RateLimiterRegularDelaySeconds specifies the rate limiter delay for regular scale objects
	// +kubebuilder:validation:Optional
	RateLimiterRegularDelaySeconds *int `json:"rateLimiterRegularDelaySeconds,omitempty"`

	// RateLimiterRegularNumOfWorkers specifies the rate limiter number of workers for regular scale objects
	// +kubebuilder:validation:Optional
	RateLimiterRegularNumOfWorkers *int `json:"rateLimiterRegularNumOfWorkers,omitempty"`

	// SyncPodsWithWorkloads specifies whether to run sync pods with the sync workloads
	// +kubebuilder:validation:Optional
	SyncPodsWithWorkloads *bool `json:"syncPodsWithWorkloads,omitempty"`

	// Replicas specifies the number of replicas of the cluster sync service
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the cluster-sync
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the cluster-sync
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (c *ClusterSync) SetDefaultsWhereNeeded(internalSpec *runv1_internal.Internal, globalReplicaCount *int32) {
	if c.Enabled == nil {
		c.Enabled = ptr.To(!internalSpec.RemoteConfig.GetSystemConfig().Deactivated)
	}

	if c.Image == nil {
		c.Image = &common.Image{}
	}
	c.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(internalSpec.Global.ImageRegistry),
		Name:     ptr.To("cluster-sync"),
	})

	if c.Resources == nil {
		c.Resources = &Resources{}
	}
	c.Resources.SetDefaultsWhereNeeded()

	if c.Metrics == nil {
		c.Metrics = &Metrics{}
	}
	c.Metrics.SetDefaultsWhereNeeded()

	if c.SSLCertificateVerification == nil {
		c.SSLCertificateVerification = ptr.To[bool](true)
	}

	if c.MaxNumOfEventsPerRequest == nil {
		c.MaxNumOfEventsPerRequest = ptr.To[int](1000)
	}

	if c.MaxNumOfPodGroupsPerRequest == nil {
		c.MaxNumOfPodGroupsPerRequest = ptr.To[int](900)
	}

	if c.MaxNumOfPodsPerRequest == nil {
		c.MaxNumOfPodsPerRequest = ptr.To[int](1800)
	}

	if c.MaxNumOfWorkloadsPerRequest == nil {
		c.MaxNumOfWorkloadsPerRequest = ptr.To[int](1000)
	}

	if c.MaxNumOfDatavolumesPerRequest == nil {
		c.MaxNumOfDatavolumesPerRequest = ptr.To[int](2000)
	}

	if c.SyncEventsIntervalSeconds == nil {
		c.SyncEventsIntervalSeconds = ptr.To[int](10)
	}

	if c.LivenessIntervalSeconds == nil {
		c.LivenessIntervalSeconds = ptr.To[int](10)
	}

	if c.MaxNumberOfRetries == nil {
		c.MaxNumberOfRetries = ptr.To[int](10)
	}

	if c.RetriesIntervalSeconds == nil {
		c.RetriesIntervalSeconds = ptr.To[int](10)
	}

	if c.SyncIntervalMinutes == nil {
		c.SyncIntervalMinutes = ptr.To[int](180)
	}
	if c.LogLevel == nil {
		c.LogLevel = ptr.To[int](int(zap.InfoLevel))
	}

	if c.ExistingSecretToSyncLabel == "" {
		c.ExistingSecretToSyncLabel = constants.RunaiResourceLabel
	}

	if c.ConfigMapToSyncLabel == "" {
		c.ConfigMapToSyncLabel = constants.RunaiResourceLabel
	}

	if c.Profiling == nil {
		c.Profiling = &Profiling{}
	}
	c.Profiling.SetDefaultsWhereNeeded()

	if c.K8sClientConfig == nil {
		c.K8sClientConfig = &K8sClientConfig{}
	}
	c.K8sClientConfig.SetDefaultsWhereNeeded()

	if c.NvidiaGPULabel == nil {
		c.NvidiaGPULabel = ptr.To[string]("feature.node.kubernetes.io/pci-10de.present")
	}

	if c.NodeLabelNvidiaContainerToolkit == nil {
		c.NodeLabelNvidiaContainerToolkit = ptr.To[string]("nvidia.com/gpu.deploy.container-toolkit")
	}

	if c.NodeLabelNvidiaDcgmExporter == nil {
		c.NodeLabelNvidiaDcgmExporter = ptr.To[string]("nvidia.com/gpu.deploy.dcgm-exporter")
	}

	if c.EnableNodeNvidiaComponentsReadiness == nil {
		c.EnableNodeNvidiaComponentsReadiness = ptr.To[bool](false)
	}

	if c.UseNewWorkloadsApi == nil {
		c.UseNewWorkloadsApi = ptr.To[bool](true)
	}

	if c.UseDeprecatedWorkloadsApi == nil {
		c.UseDeprecatedWorkloadsApi = ptr.To[bool](false)
	}

	if c.RateLimiterWorkloadsDelaySeconds == nil {
		c.RateLimiterWorkloadsDelaySeconds = ptr.To[int](10)
	}
	if c.RateLimiterWorkloadsNumOfWorkers == nil {
		c.RateLimiterWorkloadsNumOfWorkers = ptr.To[int](3)
	}
	if c.RateLimiterRegularDelaySeconds == nil {
		c.RateLimiterRegularDelaySeconds = ptr.To[int](5)
	}
	if c.RateLimiterRegularNumOfWorkers == nil {
		c.RateLimiterRegularNumOfWorkers = ptr.To[int](1)
	}
	if c.SyncPodsWithWorkloads == nil {
		c.SyncPodsWithWorkloads = ptr.To[bool](false)
	}
	if c.Replicas == nil {
		if globalReplicaCount != nil {
			c.Replicas = ptr.To[int32](*globalReplicaCount)
		} else {
			c.Replicas = ptr.To[int32](1)
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("600m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("4Gi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("300m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("1Gi")
	}
}

// Metrics defines configuration of the cluster-sync's metrics endpoint
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
