/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NodePoolNameLabel   = "runai/node-pool"
	DefaultNodePoolName = "default"

	UnexistingRunaiNodePool = "runai-unexisting-node-pool"
)

// NodePoolSpec defines the desired state of NodePool
type NodePoolSpec struct {
	// LabelKey is the label key a node needs to have to be assigned to the nodepool
	LabelKey string `json:"labelKey,omitempty"`

	// LabelValue is the label value a node needs to have to be assigned to the nodepool
	LabelValue string `json:"labelValue,omitempty"`

	// PlacementStrategy is the placement scheduler strategy
	PlacementStrategy PlacementStrategy `json:"placementStrategy,omitempty"`

	// GPUNetworkAccelerationDetection is a strategy to use for GPU Network Acceleration detection,
	// possible values are: Auto (default), Use, DontUse
	// +kubebuilder:default="Auto"
	GPUNetworkAccelerationDetection *GPUNetworkAccelerationDetection `json:"gpuNetworkAccelerationDetection,omitempty"`

	// GPUNetworkAccelerationLabelKey - Label key by which to determine GPUNetworkAccelerationDetection nodes
	// +kubebuilder:default="nvidia.com/gpu.clique"
	GPUNetworkAccelerationLabelKey string `json:"gpuNetworkAccelerationLabelKey,omitempty"`

	// MinGuaranteedRuntime - the minimum duration (e.g., 5d8h40m) that a preemptible workload is
	// guaranteed to run once scheduled and bound to a node
	MinGuaranteedRuntime string `json:"minGuaranteedRuntime,omitempty"`

	// NetworkTopologyName - the name of the Topology atached to the nodepool
	NetworkTopologyName string `json:"networkTopologyName,omitempty"`

	// TimeBasedFairShare - Time-based fairshare scheduling configuration
	TimeBasedFairShare TimeBasedFairShare `json:"timeBasedFairShare,omitempty"`

	// NumaAwareScheduling - when enabled, the Scheduler places workloads so that GPU, CPU,
	// and CPU memory are allocated within the same NUMA node where possible.
	NumaAwareScheduling bool `json:"numaAwareScheduling,omitempty"`
}

// PlacementStrategy defines the scheduling strategy of NodePool
type PlacementStrategy struct {

	// Gpu scheduling strategy for node-level placement
	Gpu string `json:"gpu,omitempty"`

	// GpuDevice scheduling strategy for device-level placement within a node
	GpuDevice string `json:"gpuDevice,omitempty"`

	// Cpu scheduling strategy
	Cpu string `json:"cpu,omitempty"`
}

type TimeBasedFairShare struct {
	// Enables time-based scheduling for the node pool. When true, the Scheduler adjusts fairshare based on historical resource usage over a defined time window, allowing past consumption to influence scheduling order. When false, scheduling is based only on the current queue state.
	Enabled bool `json:"enabled"`

	// Set the weight value that determines how a project’s historical resource usage affects its fairshare.  - 0.0: Ignore historical usage data (100% current usage).  - 1.0: Historical usage data has higher impact, it may cause fairshare to be 0.  - Values between 0.0 and 1.0 combine historical and current usage to maintain a balanced fairshare  - Values above 1 will cause historical usage to have an even higher impact on fairshare.
	HistoricalUsageWeight *float32 `json:"historicalUsageWeight,omitempty"`

	// Set the half-life time to define the decay rate of a project’s historical resource usage. The half-life defines how long it takes for past usage to diminish to half of its original influence on a project's fairshare.  - 0: disables decay; all historical usage is weighted equally.  - > 0: older usage gradually contributes less over time (e.g., with a 24-hour half-life, usage from 24 hours ago contributes 50%, from 48 hours ago 25%, and so on).
	DecayHalfLife *string `json:"decayHalfLife,omitempty" validate:"regexp=^(0|[+]?((\\\\d+(\\\\.\\\\d+)?)(m|h|d))+$)"`

	// The time period over which historical usage is evaluated. Only usage within this window contributes to fairshare. For sliding windows, this is a rolling period (e.g., 7d = always the last 7 days). For tumbling windows, this is the cycle duration before reset.
	WindowDuration *string `json:"windowDuration,omitempty" validate:"regexp=^(0|[+]?((\\\\d+(\\\\.\\\\d+)?)(m|h|d))+$)"`

	WindowType *TimeBasedFairShareWindowType `json:"windowType,omitempty"`

	// An RFC-3339 timestamp that defines the start time for all tumbling windows. Used only when windowType = Tumbling. This timestamp marks the beginning of the first fixed window, and subsequent windows repeat at intervals equal to the configured WindowDuration. For example, with a 7-day window, an anchor of 2025-01-06T08:00:00Z produces weekly windows starting every Monday at 08:00 UTC.
	TumblingWindowStartTime *metav1.Time `json:"tumblingWindowStartTime,omitempty"`
}

// TimeBasedFairShareWindowType defines the type of time window for aggregating usage data
type TimeBasedFairShareWindowType string

const (
	// TumblingWindow represents non-overlapping, fixed-size time windows
	// Example: 1-hour windows at 0-1h, 1-2h, 2-3h
	TumblingWindow TimeBasedFairShareWindowType = "tumbling"

	// SlidingWindow represents overlapping time windows that slide forward
	// Example: a 1-hour sliding window will consider the usage of the last 1 hour prior to the current time.
	SlidingWindow TimeBasedFairShareWindowType = "sliding"
)

// NodePoolStatus defines the observed state of NodePool
type NodePoolStatus struct {
	// NodePoolPhase is the recently observed phase of the nodepool.
	Phase NodePoolPhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=NodePoolPhase"`
	// A human readable message indicating details about why the nodepool is in this condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// List of Nodes that are assigned to this nodepool
	Nodes []NodeInNodePool `json:"nodes,omitempty" protobuf:"bytes,3,opt,name=nodes"`
	// Signals to all relevant parties about GPUNetworkAcceleration being detected in some nodepool's nodes
	GPUNetworkAccelerationDetected bool `json:"gpuNetworkAccelerationDetected,omitempty"`
	// Conditions is an array of current observed nodepool conditions.
	Conditions []NodePoolCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:storageversion

// NodePool is the Schema for the nodepools API
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodePoolSpec   `json:"spec,omitempty"`
	Status NodePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodePoolList contains a list of NodePool
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodePool `json:"items"`
}

// +kubebuilder:default="Auto"
// +kubebuilder:validation:Enum:=Auto;Use;DontUse
type GPUNetworkAccelerationDetection string

// These are the valid phases of nodepool.
const (
	// UseGPUNetworkAcceleration - force set GPU Network Acceleration detected
	UseGPUNetworkAcceleration GPUNetworkAccelerationDetection = "Use"
	// DontUseGPUNetworkAcceleration - force set GPU Network Acceleration not detected
	DontUseGPUNetworkAcceleration GPUNetworkAccelerationDetection = "DontUse"
	// AutoGPUNetworkAccelerationDetection - sets GPU Network Acceleration detected if any of nodepool nodes has 'nvidia.com/gpu.imex-domain' label
	AutoGPUNetworkAccelerationDetection GPUNetworkAccelerationDetection = "Auto"
)

type NodePoolPhase string

// These are the valid phases of nodepool.
const (
	// NodePoolEmpty means that no nodes are allocated for this nodepool.
	NodePoolEmpty NodePoolPhase = "Empty"
	// NodePoolUnschedulable means that all the nodes that are associated with this nodepool are not schedulable.
	NodePoolUnschedulable NodePoolPhase = "Unschedulable"
	// NodePoolDeleting means the nodepool is being deleted.
	NodePoolDeleting NodePoolPhase = "Deleting"
	// NodePoolReady means the nodepool is ready for scheduling.
	NodePoolReady NodePoolPhase = "Ready"
)

// +enum
type NodeInNodePoolStatus string

// These are the valid phases of nodepool.
const (
	// NodeUnschedulable means that the node is unschedulable.
	NodeUnschedulable NodeInNodePoolStatus = "Unschedulable"
	// NodeReady means the node is ready for scheduling.
	NodeReady NodeInNodePoolStatus = "Ready"
)

type NodeInNodePool struct {
	Name   string               `json:"name,omitempty"`
	Status NodeInNodePoolStatus `json:"status,omitempty"`
	// TopologyMismatch when true - means the node missing Nodepool Topology nodeLevels in its labels
	TopologyMismatch bool `json:"topologyMismatch,omitempty"`
}

func init() {
	SchemeBuilder.Register(&NodePool{}, &NodePoolList{})
}

type NodePoolConditionType string

const (
	// TopologyMismatch - condition True - meaning at least one of the nodes of nodepool has mismatch between network topology NodeLevels amd node labels
	NodeTopologyMismatch NodePoolConditionType = "NodeTopologyMismatch"

	NodeTopologyMismatchReason = "NodeTopologyMismatchReason"

	// ProjectReferencesExist - condition True - meaning the nodepool is being deleted but at least one
	// project still references it in its queues; deletion is blocked until all references are removed
	ProjectReferencesExist NodePoolConditionType = "ProjectReferencesExist"

	ProjectReferencesExistReason = "ProjectReferencesExistReason"
)

// NodeCondition contains condition information for a node.
type NodePoolCondition struct {
	// Type of nodepool condition.
	Type NodePoolConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=NodeConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we got an update on a given condition.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty" protobuf:"bytes,3,opt,name=lastHeartbeatTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

func (cond *NodePoolCondition) SetConditionStatusValue(isTrue bool) {
	if isTrue {
		cond.Status = v1.ConditionTrue
	} else {
		cond.Status = v1.ConditionFalse
	}
}

// setNodePoolCondition sets a condition on a nodepool's status
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
