// +kubebuilder:object:generate:=true
package runai_scheduler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

// RunAIScheduler defines the configuration of the runai-scheduler
type RunAIScheduler struct {
	// Enabled specifies if the runai-scheduler should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Args specifies the command line arguments for the scheduler
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Profiling specifies profiling configuration
	// +kubebuilder:validation:Optional
	Profiling *Profiling `json:"profiling,omitempty"`

	// Image specifies the image configuration for the runai-scheduler
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources specifies the resources configuration for the runai-scheduler
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// GOGC configures the GOGC environment variable for the scheduler container
	// +kubebuilder:validation:Optional
	GOGC *int `json:"GOGC,omitempty"`

	// Service specifies the service configuration for the runai-scheduler
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// InferencePreemptible specifies if the scheduler considers inference jobs as preemptible
	// +kubebuilder:validation:Optional
	InferencePreemptible *bool `json:"inferencePreemptible,omitempty"`

	// DetailedFitErrors configures whether the scheduler records node details in fit errors in podgroup events
	// +kubebuilder:validation:Optional
	DetailedFitErrors *bool `json:"detailedFitErrors,omitempty"`

	// AdvancedCSIScheduling configures whether the scheduler will consider CSI storage for reclaiming jobs
	// +kubebuilder:validation:Optional
	AdvancedCSIScheduling *bool `json:"advancedCSIScheduling,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the runai-scheduler
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the runai-scheduler
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *RunAIScheduler) SetDefaultsWhereNeeded() {
	if r.Enabled == nil {
		r.Enabled = ptr.To[bool](true)
	}

	if r.Resources == nil {
		r.Resources = &Resources{}
	}
	r.Resources.SetDefaultsWhereNeeded()
}

// Args defines command line arguments for the runai-scheduler
type Args struct {
	// Verbosity specifies the logging level for the runai-scheduler
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`

	// DefaultSchedulerName specifies the default name for the scheduler
	// +kubebuilder:validation:Optional
	DefaultSchedulerName *string `json:"defaultSchedulerName,omitempty"`

	// DefaultSchedulerPeriod specifies the default period for the scheduler
	// +kubebuilder:validation:Optional
	DefaultSchedulerPeriod *string `json:"defaultSchedulerPeriod,omitempty"`

	// MaxNumberConsolidationPreemptees specifies the max number of potential preemptees
	// considered in a single consolidation action
	// +kubebuilder:validation:Optional
	MaxNumberConsolidationPreemptees *int `json:"maxNumberConsolidationPreemptees,omitempty"`

	// ClientQPS specifies the qps parameter for the scheduler kube client
	// +kubebuilder:validation:Optional
	ClientQPS *int `json:"clientQPS,omitempty"`

	// ClientBurst specifies the burst parameter for the scheduler kube client
	// +kubebuilder:validation:Optional
	ClientBurst *int `json:"clientBurst,omitempty"`

	// UseSchedulingSignatures use scheduling signatures to avoid duplicate scheduling attempts for identical jobs. Defaults to false.
	// +kubebuilder:validation:Optional
	UseSchedulingSignatures *bool `json:"useSchedulingSignatures,omitempty"`

	// FullHierarchyFairness specifies whether fairness is enforced across projects and departments levels
	// +kubebuilder:validation:Optional
	FullHierarchyFairness *bool `json:"fullHierarchyFairness,omitempty"`

	// AllowConsolidatingReclaim specifies whether pipelined pods towards should count towards 'reclaimed' resources
	// +kubebuilder:validation:Optional
	AllowConsolidatingReclaim *bool `json:"allowConsolidatingReclaim,omitempty"`

	// DefaultStalenessGracePeriod specifies the default staleness grace period for jobs
	// +kubebuilder:validation:Optional
	DefaultStalenessGracePeriod *string `json:"defaultStalenessGracePeriod,omitempty"`

	// NumOfStatusRecordingWorkers specifies the max number of go routines spawned to update pod and podgroups conditions and events
	// +kubebuilder:validation:Optional
	NumOfStatusRecordingWorkers *int `json:"numOfStatusRecordingWorkers,omitempty"`
}

// Service defines configuration for the runai-scheduler service
type Service struct {
	// Type specifies the service type
	// +kubebuilder:validation:Optional
	Type *v1.ServiceType `json:"type,omitempty"`

	// Port specifies the service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the target port in the container
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

func (resources *Resources) SetDefaultsWhereNeeded() {

	if resources.Limits == nil {
		resources.Limits = v1.ResourceList{}
	}
	if resources.Requests == nil {
		resources.Requests = v1.ResourceList{}
	}

	if _, ok := resources.Limits[v1.ResourceCPU]; !ok {
		resources.Limits[v1.ResourceCPU] = resource.MustParse("2000m")
	}
	if _, ok := resources.Limits[v1.ResourceMemory]; !ok {
		resources.Limits[v1.ResourceMemory] = resource.MustParse("2048Mi")
	}

	if _, ok := resources.Requests[v1.ResourceCPU]; !ok {
		resources.Requests[v1.ResourceCPU] = resource.MustParse("0.5")
	}
	if _, ok := resources.Requests[v1.ResourceMemory]; !ok {
		resources.Requests[v1.ResourceMemory] = resource.MustParse("1024Mi")
	}
}

// Profiling defines profiling configuration for the runai-scheduler
type Profiling struct {
	// Enabled specifies whether profiling is enabled
	// +kubebuilder:validation:Optional
	Enabled bool `json:"enabled,omitempty"`
}
