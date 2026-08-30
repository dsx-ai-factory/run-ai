// +kubebuilder:object:generate:=true
package runai_container_toolkit

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

// RunAIContainerToolkit is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: the RunAI container toolkit is not reconciled in KAI-based releases.
type RunAIContainerToolkit struct {
	// +kubebuilder:validation:Optional
	Image *Image `json:"image,omitempty"`
	// +kubebuilder:validation:Optional
	InitContainerImage *common.Image `json:"init-container-image,omitempty"`
	// +kubebuilder:validation:Optional
	SidecarContainerImage *common.Image `json:"sidecar-image,omitempty"`
	// +kubebuilder:validation:Optional
	RunAISharedDirectoryHostPath *string `json:"runaiSharedDirectoryHostPath,omitempty"`
	// +kubebuilder:validation:Optional
	RunAIProcDirectoryHostPath *string `json:"runaiProcDirectoryHostPath,omitempty"`
	// +kubebuilder:validation:Optional
	ExtendedDaemon bool `json:"extendedDaemon,omitempty"`
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	AdditionalNodeAffinityMatchExpressions []corev1.NodeSelectorRequirement `json:"additionalNodeAffinityMatchExpressions,omitempty"`
	// +kubebuilder:validation:Optional
	WaitForNvidiaToolkit *bool `json:"waitForNvidiaToolkit,omitempty"`
	// +kubebuilder:validation:Optional
	Metrics *Metrics `json:"metrics,omitempty"`
	// +kubebuilder:validation:Optional
	Memory *Memory `json:"memory,omitempty"`
	// +kubebuilder:validation:Optional
	Pid *Pid `json:"pid,omitempty"`
	// +kubebuilder:validation:Optional
	Compute *Compute `json:"compute,omitempty"`
	// +kubebuilder:validation:Optional
	Swap *Swap `json:"swap,omitempty"`
	// +kubebuilder:validation:Optional
	Allocator *Allocator `json:"allocator,omitempty"`
	// +kubebuilder:validation:Optional
	Route *Route `json:"route,omitempty"`
	// +kubebuilder:validation:Optional
	NvidiaContainerRuntimeClass *bool `json:"nvidiaContainerRuntimeClass,omitempty"`
}

type Metrics struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	CollectionInterval *int `json:"collectionInterval,omitempty"`
	// +kubebuilder:validation:Optional
	CollectionWindow *int `json:"collectionWindow,omitempty"`
}

type Memory struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	OOMKiller *bool `json:"oomKiller,omitempty"`
}

type Pid struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

type Compute struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	LeaseTime *int `json:"leaseTime,omitempty"`
	// +kubebuilder:validation:Optional
	SyncTimeout *int `json:"syncTimeout,omitempty"`
	// +kubebuilder:validation:Optional
	KeepAliveTimeout *int `json:"keepAliveTimeout,omitempty"`
	// +kubebuilder:validation:Optional
	GrantedLeaseTimeout *int `json:"grantedLeaseTimeout,omitempty"`
	// +kubebuilder:validation:Optional
	Mode *string `json:"mode,omitempty"`
}

type Swap struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	PriorityReclaimer *bool `json:"priorityReclaimer,omitempty"`
	// +kubebuilder:validation:Optional
	RamLimit *string `json:"ramLimit,omitempty"`
}

type Allocator struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

type Route struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	QueuePolicy *string `json:"queuePolicy,omitempty"`
	// +kubebuilder:validation:Optional
	IdlePreemption *bool `json:"idlePreemption,omitempty"`
}

type Image common.Image
type Resources corev1.ResourceRequirements
