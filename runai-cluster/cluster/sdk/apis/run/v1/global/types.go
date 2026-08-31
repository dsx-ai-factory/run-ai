/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package global

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/utils/ptr"
)

// Global defines global configuration for the Run:AI deployment
type Global struct {
	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	RunaiContainerRuntimeEnabled *bool `json:"runaiContainerRuntimeEnabled,omitempty"`

	// SubdomainSupport configures the operator to enable subdomain urls
	// +kubebuilder:validation:Optional
	// +nullable
	SubdomainSupport *bool `json:"subdomainSupport,omitempty"`

	// WorkloadInjectedContainers configures containers that Run:AI adds to user workloads.
	// +kubebuilder:validation:Optional
	WorkloadInjectedContainers *WorkloadInjectedContainersConfig `json:"workloadInjectedContainers,omitempty"`

	// LabelUser specifies if a label containing the user name should be
	// added to jobs and pods
	// +kubebuilder:validation:Optional
	LabelUser *LabelUser `json:"labeluser,omitempty"`

	// ScheduleWorkflowPods specifies if runai should schedule workflow pods
	// +kubebuilder:validation:Optional
	ScheduleWorkflowPods *ScheduleWorkflowPods `json:"scheduleWorkflowPods,omitempty"`

	// PSPEnabled specifies if PodSecurityPolicies will be used
	// +kubebuilder:validation:Optional
	PSPEnabled bool `json:"pspEnabled,omitempty"`

	// Inference specifies if inference workloads are enabled in the cluster
	// +kubebuilder:validation:Optional
	Inference *Inference `json:"inference,omitempty"`

	// Describes the proxy settings RunAI components will use when accessing
	// services outside of the cluster
	// +kubebuilder:validation:Optional
	HTTPProxy *HTTPProxy `json:"httpProxy,omitempty"`

	// NodeAffinity specifies which nodes should accept RunAI jobs
	// +kubebuilder:validation:Optional
	NodeAffinity *NodeAffinity `json:"nodeAffinity,omitempty"`

	// SecurityContext defines security context for RunAI containers
	// +kubebuilder:validation:Optional
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`

	// OverProvisioning specifies whether job over provisioning is enabled
	// +kubebuilder:validation:Optional
	OverProvisioning *OverProvisioning `json:"overProvisioning,omitempty"`

	// IdleGPUInactiveDuration specifies how long a GPU may be idle before jobs using
	// it are marked as swappable
	// +kubebuilder:validation:Optional
	IdleGPUInactiveDuration *int `json:"idleGpuInactiveDuration,omitempty"`

	// Runtime describes the container runtime used for RunAI pods
	// +kubebuilder:validation:Optional
	Runtime string `json:"runtime,omitempty"`

	// PrometheusService describes the prometheus service in the cluster
	// +kubebuilder:validation:Optional
	PrometheusService *PrometheusService `json:"prometheusService,omitempty"`

	// +kubebuilder:validation:Optional
	NvidiaDcgmExporter *NvidiaDcgmExporter `json:"nvidiaDcgmExporter,omitempty"`

	// +kubebuilder:validation:Optional
	KubevirtCluster *KubevirtCluster `json:"kubevirtCluster,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	Core *Core `json:"core,omitempty"`

	// +kubebuilder:validation:Optional
	Ingress *Ingress `json:"ingress,omitempty"`

	// +kubebuilder:validation:Optional
	GatewayAPI *GatewayAPI `json:"gatewayAPI,omitempty"`

	// EnableWorkloadOwnershipProtection specifies if we should enable the ownership-protected workloads feature
	// +kubebuilder:validation:Optional
	EnableWorkloadOwnershipProtection *bool `json:"enableWorkloadOwnershipProtection,omitempty"`

	// ControlPlanePoliciesEnabled specifies if the agent should sync policies from the control plane
	// +kubebuilder:validation:Optional
	ControlPlanePoliciesEnabled *bool `json:"controlPlanePoliciesEnabled,omitempty"`

	// SchedulingServices specifies common configuration for scheduling services
	// +kubebuilder:validation:Optional
	SchedulingServices *SchedulingServices `json:"schedulingServices,omitempty"`

	// SyncServices specifies common configuration for sync services
	// +kubebuilder:validation:Optional
	SyncServices *SyncServices `json:"syncServices,omitempty"`

	// WorkloadServices specifies common configuration for workload services
	// +kubebuilder:validation:Optional
	WorkloadServices *WorkloadServices `json:"workloadServices,omitempty"`

	// AdditionalTolerations defines additional tolerations for runai pods that need to be deployment on all nodes
	// +kubebuilder:validation:Optional
	// Deprecated: Use 'DaemonsetsTolerations' instead
	AdditionalTolerations []v1.Toleration `json:"additionalTolerations,omitempty"`

	// ManagedNodes specifies which nodes are managed by Run:AI
	// +kubebuilder:validation:Optional
	ManagedNodes *ManagedNodes `json:"managedNodes,omitempty"`

	// ReplicaCount specifies the number of replicas of services that have no specific replicas configuration
	// +kubebuilder:validation:Optional
	ReplicaCount *int32 `json:"replicaCount,omitempty"`

	// Tolerations defines additional tolerations for runai cluster operators & services
	// +kubebuilder:validation:Optional
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// Affinity defines affinity for runai cluster operators & services
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// DaemonsetsTolerations defines additional tolerations for daemonsets in cluster
	// +kubebuilder:validation:Optional
	DaemonsetsTolerations []v1.Toleration `json:"daemonsetsTolerations,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	DevicePluginBindings *bool `json:"devicePluginBindings,omitempty"`

	// CustomCAForGit specifies custom CA configuration for connecting with git sync
	// +kubebuilder:validation:Optional
	CustomCAGit *common.CustomCA `json:"customCAGit,omitempty"`

	// CustomCAForS3 specifies custom CA configuration for connecting with S3
	// +kubebuilder:validation:Optional
	CustomCAS3 *common.CustomCA `json:"customCAS3,omitempty"`

	// RequireDefaultPodAntiAffinity specifies if the default pod anti affinity should be required and not preferred
	// +kubebuilder:validation:Optional
	RequireDefaultPodAntiAffinity bool `json:"requireDefaultPodAntiAffinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for Run:AI services
	// +kubebuilder:validation:Optional
	VPA *VPASpec `json:"vpa,omitempty"`
}

// WorkloadInjectedContainersConfig groups configuration by injected container type.
type WorkloadInjectedContainersConfig struct {
	// OAuth2Proxy configures oauth2-proxy containers used to authenticate workload endpoints.
	// +kubebuilder:validation:Optional
	OAuth2Proxy *OAuth2ProxyContainerConfig `json:"oauth2Proxy,omitempty"`

	// UsernameLoader configures username-loader init containers used by authenticated workload endpoints.
	// +kubebuilder:validation:Optional
	UsernameLoader *UsernameLoaderContainerConfig `json:"usernameLoader,omitempty"`

	// GitSync configures git-sync init containers used by workloads.
	// +kubebuilder:validation:Optional
	GitSync *GitSyncContainerConfig `json:"gitSync,omitempty"`

	// S3FileSystem configures S3 filesystem containers used by workloads.
	// +kubebuilder:validation:Optional
	S3FileSystem *S3FileSystemContainerConfig `json:"s3FileSystem,omitempty"`
}

// OAuth2ProxyContainerConfig configures oauth2-proxy containers injected into workloads.
type OAuth2ProxyContainerConfig struct {
	// Resources specifies the resource requirements for oauth2-proxy containers.
	// When omitted, oauth2-proxy containers keep their existing empty resource requirements.
	// +kubebuilder:validation:Optional
	Resources *WorkloadInjectedContainerResourceRequirements `json:"resources,omitempty"`
}

// UsernameLoaderContainerConfig configures username-loader init containers injected into workloads.
type UsernameLoaderContainerConfig struct {
	// Resources specifies the resource requirements for username-loader init containers.
	// When omitted, username-loader containers keep their existing default resource requirements.
	// +kubebuilder:validation:Optional
	Resources *WorkloadInjectedContainerResourceRequirements `json:"resources,omitempty"`
}

// GitSyncContainerConfig configures git-sync init containers injected into workloads.
type GitSyncContainerConfig struct {
	// Resources specifies the resource requirements for git-sync init containers.
	// When omitted, git-sync containers keep their existing empty resource requirements.
	// +kubebuilder:validation:Optional
	Resources *WorkloadInjectedContainerResourceRequirements `json:"resources,omitempty"`
}

// S3FileSystemContainerConfig configures S3 filesystem containers injected into workloads.
type S3FileSystemContainerConfig struct {
	// Resources specifies the resource requirements for S3 filesystem containers.
	// When omitted, S3 filesystem containers keep their existing empty resource requirements.
	// +kubebuilder:validation:Optional
	Resources *WorkloadInjectedContainerResourceRequirements `json:"resources,omitempty"`
}

// WorkloadInjectedContainerResourceRequirements specifies requests and limits for a
// container that Run:AI adds to a user workload.
type WorkloadInjectedContainerResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// +kubebuilder:validation:Optional
	Limits v1.ResourceList `json:"limits,omitempty"`

	// Requests describes the minimum amount of compute resources required.
	// +kubebuilder:validation:Optional
	Requests v1.ResourceList `json:"requests,omitempty"`
}

// Validate checks Kubernetes resource names, non-negative quantities, and
// request-to-limit relationships.
func (r *WorkloadInjectedContainerResourceRequirements) Validate() error {
	_, err := r.ToResourceRequirements()
	return err
}

// ToResourceRequirements validates and converts the RunaiConfig wire type to
// Kubernetes resource requirements.
func (r *WorkloadInjectedContainerResourceRequirements) ToResourceRequirements() (*v1.ResourceRequirements, error) {
	if r == nil {
		return nil, nil
	}
	resources := &v1.ResourceRequirements{
		Limits:   r.Limits,
		Requests: r.Requests,
	}
	if err := ValidateWorkloadInjectedContainerResourceRequirements(*resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// ValidateWorkloadInjectedContainerResourceRequirements validates the subset of
// Kubernetes resource requirements supported for Run:AI-injected containers.
func ValidateWorkloadInjectedContainerResourceRequirements(resources v1.ResourceRequirements) error {
	if err := validateWorkloadInjectedContainerResourceList("limits", resources.Limits); err != nil {
		return err
	}
	if err := validateWorkloadInjectedContainerResourceList("requests", resources.Requests); err != nil {
		return err
	}
	for name, request := range resources.Requests {
		if limit, ok := resources.Limits[name]; ok && request.Cmp(limit) > 0 {
			return fmt.Errorf("requests.%s must not exceed limits.%s", name, name)
		}
	}
	return nil
}

// ResourceRequirementsFromStringMaps parses and validates the resource maps
// used by runai-public and the cluster-service API.
func ResourceRequirementsFromStringMaps(limits, requests map[string]string) (v1.ResourceRequirements, error) {
	parsedLimits, err := resourceListFromStringMap("limits", limits)
	if err != nil {
		return v1.ResourceRequirements{}, err
	}
	parsedRequests, err := resourceListFromStringMap("requests", requests)
	if err != nil {
		return v1.ResourceRequirements{}, err
	}
	resources := v1.ResourceRequirements{Limits: parsedLimits, Requests: parsedRequests}
	if err := ValidateWorkloadInjectedContainerResourceRequirements(resources); err != nil {
		return v1.ResourceRequirements{}, err
	}
	return resources, nil
}

func resourceListFromStringMap(field string, values map[string]string) (v1.ResourceList, error) {
	if values == nil {
		return nil, nil
	}
	result := make(v1.ResourceList, len(values))
	for name, rawQuantity := range values {
		quantity, err := resource.ParseQuantity(rawQuantity)
		if err != nil {
			return nil, fmt.Errorf("%s.%s has invalid quantity %q: %w", field, name, rawQuantity, err)
		}
		result[v1.ResourceName(name)] = quantity
	}
	return result, nil
}

func validateWorkloadInjectedContainerResourceList(field string, resourceList v1.ResourceList) error {
	for name, quantity := range resourceList {
		if problems := k8svalidation.IsQualifiedName(string(name)); len(problems) > 0 {
			return fmt.Errorf("%s has invalid resource name %q: %s", field, name, problems[0])
		}
		if quantity.Sign() < 0 {
			return fmt.Errorf("%s.%s must not be negative", field, name)
		}
	}
	return nil
}

// SetDefaultsWhereNeeded sets default for unset fields
func (g *Global) SetDefaultsWhereNeeded(internalSpec *internal.Internal) {
	if g.LabelUser == nil {
		g.LabelUser = &LabelUser{}

	}
	g.LabelUser.SetDefaultsWhereNeeded()

	if g.ScheduleWorkflowPods == nil {
		g.ScheduleWorkflowPods = &ScheduleWorkflowPods{}
	}
	g.ScheduleWorkflowPods.SetDefaultsWhereNeeded()

	if g.HTTPProxy == nil {
		g.HTTPProxy = &HTTPProxy{}

	}
	g.HTTPProxy.SetDefaultsWhereNeeded()

	if g.NodeAffinity == nil {
		g.NodeAffinity = &NodeAffinity{}
	}
	g.NodeAffinity.SetDefaultsWhereNeeded()

	if g.SecurityContext == nil {
		g.SecurityContext = &SecurityContext{}
	}
	g.SecurityContext.SetDefaultsWhereNeeded()

	if g.OverProvisioning == nil {
		g.OverProvisioning = &OverProvisioning{}
	}

	if g.OverProvisioning.Swap == nil {
		g.OverProvisioning.Swap = &Swap{}
	}

	if g.OverProvisioning.Route == nil {
		g.OverProvisioning.Route = &Route{}
	}

	if g.IdleGPUInactiveDuration == nil {
		g.IdleGPUInactiveDuration = ptr.To[int](300000) // 5 minutes
	}

	if g.Runtime == "" {
		g.Runtime = "docker"
	}

	if g.PrometheusService == nil {
		g.PrometheusService = &PrometheusService{}
	}
	g.PrometheusService.SetDefaultsWhereNeeded(internalSpec)

	if g.NvidiaDcgmExporter == nil {
		g.NvidiaDcgmExporter = &NvidiaDcgmExporter{}
	}
	g.NvidiaDcgmExporter.SetDefaultsWhereNeeded()

	if g.KubevirtCluster == nil {
		g.KubevirtCluster = &KubevirtCluster{}
	}

	if g.Ingress == nil {
		g.Ingress = &Ingress{}
	}
	g.Ingress.SetDefaultsWhereNeeded()

	if g.GatewayAPI == nil {
		g.GatewayAPI = &GatewayAPI{}
	}
	g.GatewayAPI.SetDefaultsWhereNeeded()

	if g.EnableWorkloadOwnershipProtection == nil {
		g.EnableWorkloadOwnershipProtection = ptr.To[bool](false)
	}

	if g.ControlPlanePoliciesEnabled == nil {
		g.ControlPlanePoliciesEnabled = ptr.To[bool](true)
	}

	if g.SchedulingServices == nil {
		g.SchedulingServices = &SchedulingServices{}
	}

	if g.SyncServices == nil {
		g.SyncServices = &SyncServices{}
	}

	if g.WorkloadServices == nil {
		g.WorkloadServices = &WorkloadServices{}
	}

	if g.VPA == nil {
		g.VPA = &VPASpec{}
	}
	g.VPA.SetDefaultsWhereNeeded()

	if g.AdditionalTolerations == nil {
		g.AdditionalTolerations = []v1.Toleration{}
	}

	if g.ManagedNodes == nil {
		g.ManagedNodes = &ManagedNodes{}
	}
	g.ManagedNodes.SetDefaultsWhereNeeded()

	if g.ReplicaCount == nil {
		g.ReplicaCount = ptr.To[int32](1)
	}

	if g.CustomCAS3 == nil {
		g.CustomCAS3 = &common.CustomCA{}
	}
	g.CustomCAS3.SetDefaultsWhereNeeded()

	if g.CustomCAGit == nil {
		g.CustomCAGit = &common.CustomCA{}
	}
	g.CustomCAGit.SetDefaultsWhereNeeded()
}

// LabelUser defines configuration for the label-user feature
type LabelUser struct {
	// Enabled specifies if a label containing the user name should be
	// added to jobs and pods
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (labelUser *LabelUser) SetDefaultsWhereNeeded() {
	if labelUser.Enabled == nil {
		labelUser.Enabled = ptr.To[bool](true)
	}
}

// ScheduleWorkflowPods defines configuration for the schedule-workflow-pods feature
type ScheduleWorkflowPods struct {
	// Enabled specifies if runai should schedule workflow pods
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (scheduleWorkflowPods *ScheduleWorkflowPods) SetDefaultsWhereNeeded() {
	if scheduleWorkflowPods.Enabled == nil {
		scheduleWorkflowPods.Enabled = ptr.To[bool](true)
	}
}

// Inference defines configuration for the inference feature
type Inference struct {
	// Enabled specifies if inference workloads are enabled in the cluster
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}

// HTTPProxy defines configuration for the http-proxy used by Run:AI components
type HTTPProxy struct {
	// Enabled specifies if an HTTP proxy should be used for outgoing connections
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`

	// ProxyURL is the URL to the proxy server
	// +kubebuilder:validation:Optional
	ProxyURL string `json:"proxyUrl,omitempty"`

	// ConfigMapName specifies the name of the configmap that configures the proxy communication
	// +kubebuilder:validation:Optional
	ConfigMapName string `json:"configMapName,omitempty"`

	// TLSCertSecret is the reference to a secret containing the certificate used for outgoing connections
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self.name != ''",message="tlsCertSecret.name must not be empty"
	TLSCertSecret *v1.SecretKeySelector `json:"tlsCertSecret,omitempty"`

	// TLSSkipValidation specifies if unvalidated connections should be accepted
	// +kubebuilder:validation:Optional
	TLSSkipValidation bool `json:"tlsSkipValidation,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (httpProxy *HTTPProxy) SetDefaultsWhereNeeded() {
	if httpProxy.ProxyURL == "" {
		httpProxy.ProxyURL = "http://proxy-url:port"
	}

	if httpProxy.ConfigMapName == "" {
		httpProxy.ConfigMapName = "http-proxy-crt"
	}
}

// DB defines configuration for the Run:AI cluster DB
type DB struct {
	// ServerName is the name of the DB service
	// +kubebuilder:validation:Optional
	ServiceName string `json:"serviceName,omitempty"`

	// ServicePort is the port of the DB endpoint
	// +kubebuilder:validation:Optional
	ServicePort *int `json:"servicePort,omitempty"`
}

// PostgreSQL defines configuration for the postgesql database instance
type PostgreSQL struct {
	// PostgreSQLDatabase specifies the database name
	// +kubebuilder:validation:Optional
	PostgreSQLDatabase string `json:"postgresqlDatabase,omitempty"`
}

// NodeAffinity defines node affinity configurations for Run:AI components
type NodeAffinity struct {
	// RestrictScheduling prevents pure CPU workloads from being scheduled
	// on GPU nodes
	// +kubebuilder:validation:Optional
	RestrictScheduling bool `json:"restrictScheduling,omitempty"`

	// NFDGPULabel is the label to look for on GPU nodes in case nfd (nvidia-feature-discovery) is installed
	// +kubebuilder:validation:Optional
	NFDGPULabel string `json:"nfdGpuLabel,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (nodeAffinity *NodeAffinity) SetDefaultsWhereNeeded() {
	if nodeAffinity.NFDGPULabel == "" {
		nodeAffinity.NFDGPULabel = "nvidia.com/gpu.present"
	}
}

type SecurityContext v1.SecurityContext

// Swap defines configuration for the swap feature
type Swap struct {
	// Enabled specifies that the swapping feature should be enabled
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}

// Route defines configuration for the swap feature
type Route struct {
	// Enabled specifies that the swapping feature should be enabled
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}

// OverProvisioning defines configuration for the over provisioning feature
type OverProvisioning struct {
	// Swap specifies that the route capability of over provisioning feature should be enabled
	// +kubebuilder:validation:Optional
	Swap *Swap `json:"swap,omitempty"`

	// Route specifies that the route capability of over provisioning feature should be enabled
	// +kubebuilder:validation:Optional
	Route *Route `json:"route,omitempty"`
}

func (securityContext *SecurityContext) SetDefaultsWhereNeeded() {
	if securityContext.AllowPrivilegeEscalation == nil {
		securityContext.AllowPrivilegeEscalation = ptr.To[bool](false)
	}
	if securityContext.Capabilities == nil {
		securityContext.Capabilities = &v1.Capabilities{}
	}
	if len(securityContext.Capabilities.Drop) == 0 {
		securityContext.Capabilities.Drop = append(securityContext.Capabilities.Drop, "ALL")
	}
	if securityContext.RunAsNonRoot == nil {
		securityContext.RunAsNonRoot = ptr.To[bool](true)
	}
	if securityContext.SeccompProfile == nil {
		securityContext.SeccompProfile = &v1.SeccompProfile{}
	}
	if securityContext.SeccompProfile.Type == "" {
		securityContext.SeccompProfile.Type = v1.SeccompProfileTypeRuntimeDefault
	}
	if securityContext.RunAsUser == nil {
		securityContext.RunAsUser = ptr.To[int64](10000)
	}

}

// PrometheusService defines configuration for the Run:AI prometheus endpoint
type PrometheusService struct {
	// Address specifies the address of the prometheus service
	// +kubebuilder:validation:Optional
	Address string `json:"address,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (prometheusService *PrometheusService) SetDefaultsWhereNeeded(internalSpec *internal.Internal) {
	if prometheusService.Address == "" {
		if internalSpec.Global.Platform.KubernetesFlavor == internal.KubernetesFlavorOpenShift {
			prometheusService.Address = "https://thanos-querier.openshift-monitoring.svc:9091"
		} else {
			prometheusService.Address = "http://prometheus-operated.runai.svc:9090"
		}
	}
}

type NvidiaDcgmExporter struct {
	// +kubebuilder:validation:Optional
	InstalledFromGpuOperator *bool `json:"installedFromGpuOperator,omitempty"`

	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// +kubebuilder:validation:Optional
	ServiceMonitor *ServiceMonitor `json:"serviceMonitor,omitempty"`
}

func (n *NvidiaDcgmExporter) SetDefaultsWhereNeeded() {
	if n.Namespace == "" {
		n.Namespace = "gpu-operator"
	}
	if n.InstalledFromGpuOperator == nil {
		n.InstalledFromGpuOperator = ptr.To[bool](true)
	}

	if n.ServiceMonitor == nil {
		n.ServiceMonitor = &ServiceMonitor{}
	}
	n.ServiceMonitor.SetDefaultsWhereNeeded()
}

type ServiceMonitor struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Optional
	Interval string `json:"interval,omitempty"`
}

func (serviceMonitor *ServiceMonitor) SetDefaultsWhereNeeded() {
	if serviceMonitor.Enabled == nil {
		serviceMonitor.Enabled = ptr.To[bool](true)
	}
}

type KubevirtCluster struct {
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}

// Core is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: Core configuration is not consumed in KAI-based releases.
type Core struct {
	// +kubebuilder:validation:Optional
	Analytics *Analytics `json:"analytics,omitempty"`
	// +kubebuilder:validation:Optional
	Debug *Debug `json:"debug,omitempty"`
	// +kubebuilder:validation:Optional
	TimeSlicing *TimeSlicing `json:"timeSlicing,omitempty"`
	// +kubebuilder:validation:Optional
	ToolkitEvents *ToolkitEvents `json:"toolkitEvents,omitempty"`
}

type Analytics struct {
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}

type Debug struct {
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Level string `json:"level,omitempty"`
	// +kubebuilder:validation:Optional
	File string `json:"file,omitempty"`
	// +kubebuilder:validation:Optional
	HostPath string `json:"hostPath,omitempty"`
}

type TimeSlicing struct {
	// +kubebuilder:validation:Optional,Minimum=1,Maximum=25
	Granularity *int `json:"granularity,omitempty"`
	// +kubebuilder:validation:Optional,Minimum=1
	LeaseTime *int `json:"leaseTime,omitempty"`
	// +kubebuilder:validation:Optional,Minimum=0
	RateLimitCount *int `json:"rateLimitCount,omitempty"`
}

// ToolkitEventReason is the local wire-compatible form of the former Engine SDK type.
// +kubebuilder:validation:Enum=gpuOOMKilled;unresponsive;nodeScheduler
type ToolkitEventReason string

type ToolkitEvents struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Mask *[]ToolkitEventReason `json:"mask,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

type SchedulingServices struct {
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for scheduling services
	// +kubebuilder:validation:Optional
	VPA *VPASpec `json:"vpa,omitempty"`
}

type SyncServices struct {
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for sync services
	// +kubebuilder:validation:Optional
	VPA *VPASpec `json:"vpa,omitempty"`
}

type WorkloadServices struct {
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for workload services
	// +kubebuilder:validation:Optional
	VPA *VPASpec `json:"vpa,omitempty"`
}

// VPASpec defines Vertical Pod Autoscaler configuration
type VPASpec struct {
	// Enabled specifies if VPA should be enabled
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// UpdatePolicy controls when and how VPA applies changes to pod resources
	// +kubebuilder:validation:Optional
	UpdatePolicy *vpav1.PodUpdatePolicy `json:"updatePolicy,omitempty"`

	// ResourcePolicy controls how VPA computes recommended resources for containers
	// +kubebuilder:validation:Optional
	ResourcePolicy *vpav1.PodResourcePolicy `json:"resourcePolicy,omitempty"`
}

func (v *VPASpec) SetDefaultsWhereNeeded() {
	if v.Enabled == nil {
		v.Enabled = ptr.To(false)
	}
	if v.UpdatePolicy == nil {
		mode := vpav1.UpdateModeInPlaceOrRecreate
		v.UpdatePolicy = &vpav1.PodUpdatePolicy{
			UpdateMode: &mode,
		}
	}
}

type Ingress struct {
	// Enabled specifies whether Ingress resources should be created.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Annotations allows to configure the ingress according to the ingress-class specifications, defaults for nginx.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Class specifies the ingress-class to be used for this ingress, defaults to nginx.
	Class string `json:"ingressClass,omitempty"`
}

func (ingress *Ingress) SetDefaultsWhereNeeded() {
	if ingress.Enabled == nil {
		ingress.Enabled = ptr.To(true)
	}
	if ingress.Class == "" {
		ingress.Class = "nginx"
	}
}

// +kubebuilder:validation:XValidation:rule="!has(self.enabled) || !self.enabled || (has(self.name) && self.name != '' && has(self.__namespace__) && self.__namespace__ != '')",message="name and namespace are required when gatewayAPI is enabled"
type GatewayAPI struct {
	// Enabled specifies whether Gateway API HTTPRoute resources should be created.
	Enabled *bool `json:"enabled,omitempty"`

	// Name is the name of the Gateway resource to reference in parentRefs.
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the Gateway resource.
	Namespace string `json:"namespace,omitempty"`
}

func (g *GatewayAPI) SetDefaultsWhereNeeded() {
	if g.Enabled == nil {
		g.Enabled = ptr.To(false)
	}
}

type ManagedNodes struct {
	// InclusionCriteria specifies the criteria for nodes to be managed by Run:AI.
	// If not specified, all nodes in the cluster are managed.
	// +kubebuilder:validation:Optional
	InclusionCriteria *v1.NodeSelector `json:"inclusionCriteria,omitempty"`
}

func (managedNodes *ManagedNodes) SetDefaultsWhereNeeded() {
	if managedNodes.InclusionCriteria == nil {
		managedNodes.InclusionCriteria = &v1.NodeSelector{}
	}

	if managedNodes.InclusionCriteria.NodeSelectorTerms == nil {
		managedNodes.InclusionCriteria.NodeSelectorTerms = []v1.NodeSelectorTerm{}
	}
}
