/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GPUNetworkAccelerationDetection controls how GPU network acceleration is detected for a NodePool.
// +kubebuilder:validation:Enum=Auto;Use;DontUse
type GPUNetworkAccelerationDetection string

const (
	// AutoGPUNetworkAccelerationDetection detects GPU network acceleration via the label key on the node.
	AutoGPUNetworkAccelerationDetection GPUNetworkAccelerationDetection = "Auto"
	// UseGPUNetworkAcceleration forces GPU network acceleration to be detected.
	UseGPUNetworkAcceleration GPUNetworkAccelerationDetection = "Use"
	// DontUseGPUNetworkAcceleration forces GPU network acceleration to not be detected.
	DontUseGPUNetworkAcceleration GPUNetworkAccelerationDetection = "DontUse"
)

const (
	// Annotations on the NodePool CR configuring GPU network acceleration detection and its result.
	// Valid values for AnnotationGPUNetworkAccelerationDetection are defined by GPUNetworkAccelerationDetection;
	// absent means detection is disabled.
	AnnotationGPUNetworkAccelerationDetection = "kai/gpu-network-acceleration-detection"
	// AnnotationGPUNetworkAccelerationLabelKey is an annotation on the NodePool CR that specifies the label key used to detect GPU network acceleration on nodes.
	AnnotationGPUNetworkAccelerationLabelKey = "kai/gpu-network-acceleration-label-key"
	AnnotationGPUNetworkAccelerationDetected = "kai/gpu-network-acceleration-detected" // controller-written

	// DefaultGPUNetworkAccelerationLabelKey is the default label key for AnnotationGPUNetworkAccelerationDetection.
	// It is used by the controller to detect GPU network acceleration on nodes if the annotation is not set
	// and the detection mode is set to Auto.
	DefaultGPUNetworkAccelerationLabelKey = "nvidia.com/gpu.clique"

	// AnnotationTopologyManagerPolicy is written on a Node to expose its Kubelet Topology Manager policy.
	AnnotationTopologyManagerPolicy = "kai.resources/topology-manager-policy"
)

// NodePoolSpec defines the desired state of NodePool.
// labelKey and labelValue identify the pool's nodes and cannot change after
// creation, including by adding or dropping a field.
// +kubebuilder:validation:XValidation:rule="has(self.labelKey) == has(oldSelf.labelKey) && (!has(self.labelKey) || self.labelKey == oldSelf.labelKey) && has(self.labelValue) == has(oldSelf.labelValue) && (!has(self.labelValue) || self.labelValue == oldSelf.labelValue)",message="labelKey and labelValue are immutable"
type NodePoolSpec struct {
	// LabelKey is the label key a node must have to be assigned to this nodepool.
	LabelKey string `json:"labelKey,omitempty"`

	// LabelValue is the label value a node must have to be assigned to this nodepool.
	LabelValue string `json:"labelValue,omitempty"`

	// PreferredNetworkTopologyName is the name of the NetworkTopology attached to the nodepool.
	PreferredNetworkTopologyName string `json:"preferredNetworkTopologyName,omitempty"`

	// SchedulingShardConfig holds all per-shard scheduler settings for this NodePool.
	// Unset fields fall back to cluster SchedulerDefaults, then KAI SetDefaultsWhereNeeded.
	// +optional
	SchedulingShardConfig *SchedulingShardConfig `json:"schedulingShardConfig,omitempty"`
}

// SchedulingShardConfig defines the scheduler configuration for a NodePool's scheduling shard.
// It exposes the same surface as kai.scheduler/v1.SchedulingShardSpec, minus the
// controller-managed fields (partitionLabelValue, kValue, usageDBConfig), and adds
// timeBasedFairShare as a user-friendly replacement for kValue + usageDBConfig.
type SchedulingShardConfig struct {
	// Args are free-form CLI flags passed to the scheduler (e.g. verbosity, QPS, CSI flags).
	// Keys match the scheduler's flag names exactly.
	// +optional
	Args map[string]string `json:"args,omitempty"`

	// PlacementStrategy is the GPU/CPU bin-pack or spread strategy for this shard.
	// +optional
	PlacementStrategy *kaiv1.PlacementStrategy `json:"placementStrategy,omitempty"`

	// QueueDepthPerAction is the maximum number of jobs tried per action per queue.
	// +optional
	QueueDepthPerAction map[string]int `json:"queueDepthPerAction,omitempty"`

	// MinRuntime specifies the minimum guaranteed runtime before a job can be preempted or reclaimed.
	// +optional
	MinRuntime *kaiv1.MinRuntime `json:"minRuntime,omitempty"`

	// Plugins allows overriding plugin configuration. Keys are plugin names.
	// +optional
	Plugins map[string]kaiv1.PluginConfig `json:"plugins,omitempty"`

	// Actions allows overriding action configuration. Keys are action names.
	// +optional
	Actions map[string]kaiv1.ActionConfig `json:"actions,omitempty"`

	// TimeBasedFairShare configures usage-aware fairness for the nodepool.
	// The controller compiles this into the shard's kValue + usageDBConfig.
	// +optional
	TimeBasedFairShare *TimeBasedFairShare `json:"timeBasedFairShare,omitempty"`
}

// TimeBasedFairShare configures usage-aware fairness for the nodepool.
// The controller compiles this into the scheduling shard's kValue + usageDBConfig.usageParams.
// Not a 1:1 mapping; some fields are combined or renamed for clarity.
type TimeBasedFairShare struct {
	// Enabled activates usage-aware fairness. When false, scheduling uses current
	// queue state only.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// HistoricalUsageWeight determines how strongly past usage lowers current fair-share.
	// 0 = ignore history; 1.0 = strong; >1 = stronger. Maps to shard kValue.
	// +optional
	// +kubebuilder:validation:Minimum=0
	HistoricalUsageWeight *float32 `json:"historicalUsageWeight,omitempty"`

	// HalfLifePeriod is the half-life of past usage. 0 = no decay (all history weighted equally).
	// Format: duration string such as "24h", "7d", "0".
	// +optional
	// +kubebuilder:validation:Pattern=`^(0|([0-9]+(\.[0-9]+)?(m|h|d))+)$`
	HalfLifePeriod *string `json:"halfLifePeriod,omitempty"`

	// Window defines the time window over which usage is evaluated.
	// +optional
	Window *FairShareWindow `json:"window,omitempty"`

	// Sampling configures advanced usage-sampling tuning. Defaults are sensible;
	// most users leave this unset.
	// +optional
	Sampling *FairShareSampling `json:"sampling,omitempty"`
}

// FairShareWindow defines the time window used for usage aggregation.
// It maps to shard usageDBConfig.usageParams.windowType and windowSize.
// Not a 1:1 mapping; some fields are combined or renamed for clarity.
type FairShareWindow struct {
	// Type is the window aggregation style.
	// sliding: rolling period. tumbling: fixed resetting buckets. cron: custom schedule.
	// +optional
	// +kubebuilder:validation:Enum=sliding;tumbling;cron
	Type *string `json:"type,omitempty"`

	// Size is the window length. sliding: rolling period (e.g. "7d").
	// tumbling: cycle length. Default 7d.
	// +optional
	// +kubebuilder:validation:Pattern=`^(0|([0-9]+(\.[0-9]+)?(m|h|d))+)$`
	Size *string `json:"size,omitempty"`

	// TumblingStartTime is the anchor timestamp for tumbling windows (RFC-3339).
	// Only used when Type=tumbling.
	// +optional
	TumblingStartTime *metav1.Time `json:"tumblingStartTime,omitempty"`

	// CronString is a cron expression defining the window. Only used when Type=cron.
	// +optional
	CronString *string `json:"cronString,omitempty"`
}

// FairShareSampling configures how usage data is fetched.
// It maps to shard usageDBConfig.usageParams.{fetchInterval, stalenessPeriod, waitTimeout}.
// Not a 1:1 mapping; some fields are combined or renamed for clarity.
type FairShareSampling struct {
	// FetchInterval is how often usage is fetched. Default 1m.
	// +optional
	FetchInterval *string `json:"fetchInterval,omitempty"`

	// StalenessPeriod is when fetched usage is considered stale. Default 5m.
	// +optional
	StalenessPeriod *string `json:"stalenessPeriod,omitempty"`

	// WaitTimeout is the fetch wait timeout. Default 1m.
	// +optional
	WaitTimeout *string `json:"waitTimeout,omitempty"`
}

// NodePoolStatus defines the observed state of NodePool.
type NodePoolStatus struct {
	// Phase is the recently observed lifecycle phase of the nodepool.
	Phase NodePoolPhase `json:"phase,omitempty"`

	// Message is a human-readable description of the current phase.
	// +optional
	Message string `json:"message,omitempty"`

	// Nodes lists the nodes assigned to this nodepool and their individual status.
	// +optional
	Nodes []NodeInNodePool `json:"nodes,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// Conditions is an array of current observed nodepool conditions.
	// +optional
	Conditions []NodePoolCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=np
//+kubebuilder:storageversion

// NodePool is the Schema for the nodepools.kai API.
// A NodePool is cluster-scoped and its default instance ("default") must always exist.
// The finalizer kai.finalizers.nodepool is managed by the controller.
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec   `json:"spec,omitempty"`
	Status NodePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodePoolList contains a list of NodePool.
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
}

// NodePoolPhase represents the lifecycle phase of a nodepool.
type NodePoolPhase string

const (
	// NodePoolEmpty means no nodes are allocated for this nodepool.
	NodePoolEmpty NodePoolPhase = "Empty"
	// NodePoolUnschedulable means all nodes associated with this nodepool are not schedulable.
	NodePoolUnschedulable NodePoolPhase = "Unschedulable"
	// NodePoolDeleting means the nodepool is being deleted.
	NodePoolDeleting NodePoolPhase = "Deleting"
	// NodePoolReady means the nodepool is ready for scheduling.
	NodePoolReady NodePoolPhase = "Ready"
	// NodePoolMissingPrerequisites means the nodepool has nodes that do not meet
	// the prerequisites required for one or more enabled scheduling capabilities.
	NodePoolMissingPrerequisites NodePoolPhase = "MissingPrerequisites"
)

// NodeInNodePoolStatus is the status of an individual node within a nodepool.
type NodeInNodePoolStatus string

const (
	// NodeUnschedulable means the node is not schedulable.
	NodeUnschedulable NodeInNodePoolStatus = "Unschedulable"
	// NodeReady means the node is ready for scheduling.
	NodeReady NodeInNodePoolStatus = "Ready"
	// NodeMissingNrtHealthyPrerequisite means the node belongs to a NUMA-aware nodepool
	// but its NodeResourceTopology (NRT) data is missing or its NUMA configuration is
	// invalid, so NUMA-aware scheduling cannot be honored on this node.
	NodeMissingNrtHealthyPrerequisite NodeInNodePoolStatus = "MissingNrtHealthyPrerequisite"
)

// NodeInNodePool describes an individual node that is part of a NodePool.
type NodeInNodePool struct {
	Name   string               `json:"name,omitempty"`
	Status NodeInNodePoolStatus `json:"status,omitempty"`
	// TopologyMismatch is true when the node is missing NetworkTopology node-level labels.
	TopologyMismatch bool `json:"topologyMismatch,omitempty"`
}

// NodePoolConditionType identifies the type of a NodePool condition.
type NodePoolConditionType string

const (
	// NodeTopologyMismatch indicates that at least one node has a mismatch between
	// the NetworkTopology NodeLevels and its labels.
	NodeTopologyMismatch NodePoolConditionType = "NodeTopologyMismatch"

	// NodeTopologyMismatchReason is the condition reason set when a NodeTopologyMismatch
	// is detected on one or more of the nodepool's nodes.
	NodeTopologyMismatchReason = "NodeTopologyMismatchReason"

	// ProjectReferencesExist indicates that projects still reference this nodepool,
	// blocking deletion.
	ProjectReferencesExist NodePoolConditionType = "ProjectReferencesExist"

	// ProjectReferencesExistReason is the condition reason set when the nodepool's
	// deletion is blocked by projects still referencing it.
	ProjectReferencesExistReason = "ProjectReferencesExistReason"

	// NodePoolMissingNrtHealthyPrerequisite indicates that at least one node in a
	// NUMA-aware nodepool has unhealthy NodeResourceTopology data.
	NodePoolMissingNrtHealthyPrerequisite NodePoolConditionType = "MissingNrtHealthyPrerequisite"

	// NodePoolMissingNrtHealthyPrerequisiteReason is the condition reason set when
	// unhealthy NodeResourceTopology data is detected in the nodepool.
	NodePoolMissingNrtHealthyPrerequisiteReason = "MissingNrtHealthyPrerequisiteReason"
)

// NodePoolCondition contains condition information for a NodePool.
type NodePoolCondition struct {
	// Type of nodepool condition.
	Type NodePoolConditionType `json:"type"`

	// Status of the condition (True, False, Unknown).
	Status v1.ConditionStatus `json:"status"`

	// LastHeartbeatTime is when the condition was last updated.
	// +optional
	// +nullable
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`

	// LastTransitionTime is when the condition last changed status.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine-readable reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable description of the condition's last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// SetConditionStatusValue sets the condition status to True or False based on isTrue.
func (cond *NodePoolCondition) SetConditionStatusValue(isTrue bool) {
	if isTrue {
		cond.Status = v1.ConditionTrue
	} else {
		cond.Status = v1.ConditionFalse
	}
}

// SetNodePoolCondition sets a condition on a nodepool's status.
func (nodePool *NodePool) SetNodePoolCondition(condition NodePoolCondition) {
	for i := range nodePool.Status.Conditions {
		existingCondition := nodePool.Status.Conditions[i]
		if existingCondition.Type == condition.Type {
			if existingCondition.Status == condition.Status {
				condition.LastTransitionTime = existingCondition.LastTransitionTime
			} else {
				condition.LastTransitionTime = metav1.Now()
			}
			nodePool.Status.Conditions[i] = condition
			return
		}
	}

	if condition.Status == v1.ConditionTrue {
		condition.LastTransitionTime = metav1.Now()
		nodePool.Status.Conditions = append(nodePool.Status.Conditions, condition)
	}
}
