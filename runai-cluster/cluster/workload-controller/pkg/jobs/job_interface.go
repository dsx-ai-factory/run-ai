package jobs

import (
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	core1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobAdapter interface {
	client.Object
	GetOriginalObject() client.Object
	GetSpec() Spec
	GetStatus() JobStatus
	GetJobKind() string
}

type Spec interface {
	GetReplicaTypeSpecs() map[constants.ReplicaType]ReplicaTypeSpec
	GetReplicaSpec(replicaType constants.ReplicaType) ReplicaTypeSpec
}

type ReplicaTypeSpec interface {
	GetPodTemplateSpec() *core1.PodTemplateSpec
	GetReplicas() *int32
}

type JobAdapterFactory interface {
	CreateJobAdapter(fromObject client.Object) (JobAdapter, error)
}

type JobStatus interface {
	// IsJobFinished - check if the job is finished
	IsJobFinished() bool
	// IsTerminating - check if the job is termination
	IsTerminating() bool
	// IsJobFailed - check if the job is failed
	IsJobFailed() bool
	// IsJobSucceeded - check if the job is succeeded
	IsJobSucceeded() bool
	// GetCompletionTime - get the completion time of the job
	GetCompletionTime() *v1.Time
	// GetReadyPods - get the number of ready pods, which are pods with ready condition
	GetReadyPods() int32
	//GetRunningPods - get the number of running pods
	GetRunningPods() int32
	// GetSucceededPods - get the number of succeeded pods
	GetSucceededPods() int32
	// GetFailedPods - get the failed number of pods
	GetFailedPods() int32
	// GetPendingPods - get the pending number of pods
	GetPendingPods() int32
	// IsPending - check if the job is pending
	IsPending() bool
	// IsRunning - check if the job is running
	IsRunning() bool
	// IsInitializing - check if job is initializing
	IsInitializing() bool
	// IsCreating - check if the job is creating
	IsCreating() bool
	// IsSuspended - check if the job is suspended
	IsSuspended() bool
	// GetSuspendInfo - get the suspend info
	GetSuspendInfo() string
	// GetFailedInfo - get the failed info
	GetFailedInfo() string
	// GetAllocatedGPU - get the allocated GPU
	GetAllocatedGPU() string
	// GetAllocatedGPUMemory - get the allocated GPU memory
	GetAllocatedGPUMemory() int64
}
