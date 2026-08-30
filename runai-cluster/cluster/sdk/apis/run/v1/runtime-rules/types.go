// +kubebuilder:object:generate:=true
package runtime_rules

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// RuntimeRulesConfiguration defines configuration for the Run:AI runtime-rules component
type RuntimeRulesConfiguration struct {
	// Enabled specifies if runtime-rules should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// DisableWorkloadRunnerRules turns off runtime-rules enforcement for WorkloadRunners that wrap workloads cast from v1 to v2.
	// Defaults to false, meaning the runtime-rules watches WorkloadRunner CRDs and enforces project scheduling-rules on them. Set true to opt out.
	// +kubebuilder:validation:Optional
	DisableWorkloadRunnerRules *bool `json:"disableWorkloadRunnerRules,omitempty"`

	// Image is the configuration of the runtime-rules image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the runtime-rules pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// LogLevel specifies the log level of the runtime-rules (-1=debug, 0=info, 1=warning, 2=error...)
	// +kubebuilder:validation:Optional
	LogLevel *int `json:"logLevel,omitempty"`

	/// WorkerSleepSeconds specifies the sleep time for each worker in seconds
	// +kubebuilder:validation:Optional
	WorkerSleepSeconds *int `json:"workerSleepSeconds,omitempty"`

	/// NumberOfWorkersPerQueue specifies the number of workers per queue
	// +kubebuilder:validation:Optional
	NumberOfWorkersPerQueue *int `json:"numberOfWorkersPerQueue,omitempty"`

	/// NumberOfWorkersAutoDeleteCompleted specifies the number of workers in the AutoDeleteCompleted controller
	// +kubebuilder:validation:Optional
	NumberOfWorkersAutoDeleteCompleted *int `json:"numberOfWorkersAutoDeleteCompleted,omitempty"`

	/// DistributedFrameworkStatusPollingIntervalSeconds specifies the interval in seconds between the polling of distributed frameworks status from the config
	// +kubebuilder:validation:Optional
	DistributedFrameworkStatusPollingIntervalSeconds *int `json:"distributedFrameworkStatusPollingIntervalSeconds,omitempty"`

	// EventList Specify the notification event that will be created (integer list between 1-99%), empty list means
	// no notification at all.
	// +kubebuilder:validation:Optional
	EventList *[]int `json:"eventList,omitempty"`

	// Replicas specifies the number of replicas of the runtime-rules service
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the runtime-rules
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the runtime-rules
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (c *RuntimeRulesConfiguration) SetDefaultsWhereNeeded(globalRegistry string, globalReplicaCount *int32) {
	if c.Enabled == nil {
		c.Enabled = ptr.To[bool](true)
	}

	// DisableWorkloadRunnerRules is intentionally not defaulted here: leaving it nil (false) means
	// WorkloadRunner rule enforcement is enabled by default; the operator only passes the flag when it is explicitly true.

	if c.Image == nil {
		c.Image = &common.Image{}
	}
	c.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("runtime-rules"),
	})
	if c.Resources == nil {
		c.Resources = &Resources{}
	}
	c.Resources.SetDefaultsWhereNeeded()

	if c.LogLevel == nil {
		c.LogLevel = ptr.To[int](int(zap.InfoLevel))
	}

	if c.WorkerSleepSeconds == nil {
		c.WorkerSleepSeconds = ptr.To[int](5)
	}

	if c.NumberOfWorkersPerQueue == nil {
		c.NumberOfWorkersPerQueue = ptr.To[int](20)
	}

	if c.NumberOfWorkersAutoDeleteCompleted == nil {
		c.NumberOfWorkersAutoDeleteCompleted = ptr.To[int](4)
	}

	if c.DistributedFrameworkStatusPollingIntervalSeconds == nil {
		c.DistributedFrameworkStatusPollingIntervalSeconds = ptr.To[int](60)
	}

	if c.EventList == nil {
		c.EventList = &[]int{50}
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
		r.Limits[v1.ResourceCPU] = resource.MustParse("200m")
	}

	if _, exists := r.Limits[v1.ResourceMemory]; !exists {
		r.Limits[v1.ResourceMemory] = resource.MustParse("512Mi")
	}

	if _, exists := r.Requests[v1.ResourceCPU]; !exists {
		r.Requests[v1.ResourceCPU] = resource.MustParse("100m")
	}

	if _, exists := r.Requests[v1.ResourceMemory]; !exists {
		r.Requests[v1.ResourceMemory] = resource.MustParse("256Mi")
	}
}
