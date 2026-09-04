/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"fmt"
	clusterCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	statusUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RunaiJobAdapter struct {
	*runaiJobV1.RunaiJob
	status *RunaiJobStatus
	spec   *RunaiJobSpec
}

func NewRunaiJobAdapter(runaiJob *runaiJobV1.RunaiJob) *RunaiJobAdapter {
	spec := &RunaiJobSpec{RunaiJob: runaiJob}
	status := &RunaiJobStatus{RunaiJob: runaiJob}

	return &RunaiJobAdapter{RunaiJob: runaiJob, spec: spec, status: status}
}

func (rja *RunaiJobAdapter) GetOriginalObject() client.Object {
	return rja.RunaiJob
}

func (rja *RunaiJobAdapter) GetJobKind() string {
	return clusterCommon.KindRunaiJob
}

func (rja *RunaiJobAdapter) GetSpec() Spec {
	if rja.spec == nil {
		rja.spec = &RunaiJobSpec{RunaiJob: rja.RunaiJob}
	}
	return rja.spec
}

func (rja *RunaiJobAdapter) GetStatus() JobStatus {
	if rja.status == nil {
		rja.status = NewRunaiJobStatus(rja.RunaiJob)
	}
	return rja.status
}

func (rja *RunaiJobAdapter) GetConcurrentJob() client.Object {
	return rja.RunaiJob
}

type RunaiJobSpec struct {
	*runaiJobV1.RunaiJob
}

func (self *RunaiJobSpec) GetPodTemplateSpec() *v1.PodTemplateSpec {
	return &self.RunaiJob.Spec.Template
}

func (self *RunaiJobSpec) GetReplicas() *int32 {
	return self.Spec.Parallelism
}

func (self *RunaiJobSpec) GetReplicaTypeSpecs() map[constants.ReplicaType]ReplicaTypeSpec {
	return map[constants.ReplicaType]ReplicaTypeSpec{constants.Worker: self}
}

func (self *RunaiJobSpec) GetReplicaSpec(replicaType constants.ReplicaType) ReplicaTypeSpec {
	if replicaType == constants.Worker {
		return self
	}
	return nil
}

type RunaiJobStatus struct {
	*runaiJobV1.RunaiJob
	cachedSuspendCondition *runaiJobV1.JobCondition
	cachedFailedCondition  *runaiJobV1.JobCondition
	cachedFinishTime       *metav1.Time
}

func NewRunaiJobStatus(runaiJob *runaiJobV1.RunaiJob) *RunaiJobStatus {
	return &RunaiJobStatus{RunaiJob: runaiJob}
}

func (rjs *RunaiJobStatus) GetConcurrentJob() client.Object {
	return rjs.RunaiJob
}

func (rjs *RunaiJobStatus) IsJobFinished() bool {
	return rjs.RunaiJob != nil && (!rjs.Status.CompletionTime.IsZero() || rjs.IsJobFailed())
}

func (rjs *RunaiJobStatus) IsTerminating() bool {
	return rjs.RunaiJob != nil && !rjs.DeletionTimestamp.IsZero()
}

func (rjs *RunaiJobStatus) IsJobSucceeded() bool {
	if rjs.RunaiJob == nil {
		return false
	}

	if rjs.IsJobFinished() {
		return !rjs.IsJobFailed()
	}
	return false
}

func (rjs *RunaiJobStatus) IsJobFailed() bool {
	if rjs.RunaiJob == nil {
		return false
	}

	// if cachedFailedCondition is not nil, return the cached value
	if rjs.cachedFailedCondition != nil {
		return rjs.cachedFailedCondition.Status == v1.ConditionTrue
	}
	failed := statusUtils.GetRunaiCondition(rjs.Status.Conditions, runaiJobV1.JobFailed)
	// if failed is not nil, update the cached value and return it
	if failed != nil {
		rjs.cachedFailedCondition = failed
	} else {
		rjs.cachedFailedCondition = &runaiJobV1.JobCondition{
			Type:   runaiJobV1.JobFailed,
			Status: v1.ConditionFalse,
		}
	}
	return rjs.cachedFailedCondition.Status == v1.ConditionTrue
}

func (rjs *RunaiJobStatus) GetCompletionTime() *metav1.Time {
	if !rjs.IsJobFinished() {
		return nil
	}
	// if completion time is nil but the job completed, means the job failed
	// currently we don't have a way to know the exact time the job failed
	// runai controller doesn't update the completion time when the job fails
	if rjs.IsJobFailed() {
		if rjs.cachedFinishTime != nil {
			return rjs.cachedFinishTime
		}
		condition := statusUtils.GetRunaiCondition(rjs.Status.Conditions, runaiJobV1.JobFailed)
		if condition != nil && condition.Status == v1.ConditionTrue {
			if condition.LastTransitionTime.IsZero() {
				rjs.cachedFinishTime = &metav1.Time{Time: time.Now()}
			} else {

				rjs.cachedFinishTime = condition.LastTransitionTime.DeepCopy()
			}
			return rjs.cachedFinishTime
		}
		// if the job failed but the condition is not updated yet, return nil
		return nil
	}
	return rjs.Status.CompletionTime.DeepCopy()
}

// GetSuspendReason returns the reason for the job being suspended
func (rjs *RunaiJobStatus) GetSuspendInfo() string {
	if rjs.IsSuspended() {
		if rjs.cachedSuspendCondition != nil {
			return rjs.cachedSuspendCondition.Message
		}
	}
	return constants.EmptyStr

}

func (rjs *RunaiJobStatus) GetAllocatedGPU() string {
	if allocatedGpu, ok := rjs.Annotations[common.RunaiCurrentAllocatedGpusAnnotation]; ok {
		return allocatedGpu
	}
	return constants.EmptyStr
}

func (rjs *RunaiJobStatus) GetAllocatedGPUMemory() int64 {
	if memoryString, ok := rjs.Annotations[common.RunaiCurrentAllocatedGpuMemoryAnnotation]; ok {
		memory, err := strconv.ParseInt(memoryString, 10, 64)
		if err == nil {
			return memory
		}
	}
	return 0
}

func (rjs *RunaiJobStatus) IsSuspended() bool {
	if rjs.RunaiJob == nil {
		return false
	}

	// if we already have the suspend cached condition, return it
	if rjs.cachedSuspendCondition != nil {
		return rjs.cachedSuspendCondition.Status == v1.ConditionTrue
	}

	suspendCondition := statusUtils.GetRunaiCondition(rjs.Status.Conditions, runaiJobV1.JobSuspended)
	// for performance reasons, we cache the suspend condition
	if suspendCondition != nil {
		rjs.cachedSuspendCondition = suspendCondition
	} else {
		rjs.cachedSuspendCondition = &runaiJobV1.JobCondition{
			Type:   runaiJobV1.JobSuspended,
			Status: v1.ConditionFalse,
		}
	}
	return rjs.cachedSuspendCondition.Status == v1.ConditionTrue
}

func (rjs *RunaiJobStatus) IsPending() bool {
	if rjs.RunaiJob != nil && !rjs.GetCompletionTime().IsZero() {
		return false
	}

	return rjs.GetPendingPods() > 0
}

func (rjs *RunaiJobStatus) GetFailedInfo() string {
	if rjs.IsJobFailed() {
		if rjs.cachedFailedCondition != nil {
			return rjs.cachedFailedCondition.Message
		}
	}
	return constants.EmptyStr
}

// GetReadyPods - get the number of ready pods, which are pods with ready condition
func (rjs *RunaiJobStatus) GetReadyPods() int32 {
	// workaround bc the runai status is incorrect
	if rjs.RunaiJob != nil {
		return rjs.RunaiJob.Status.Ready
	}
	return 0
}

func (rjs *RunaiJobStatus) GetRunningPods() int32 {
	if rjs.RunaiJob != nil {
		return rjs.RunaiJob.Status.Running
	}
	return 0
}

func (rjs *RunaiJobStatus) GetSucceededPods() int32 {
	return rjs.Status.Succeeded
}

func (rjs *RunaiJobStatus) GetFailedPods() int32 {
	return rjs.Status.Failed
}

func (rjs *RunaiJobStatus) GetPendingPods() int32 {
	if rjs.RunaiJob != nil {
		return rjs.RunaiJob.Status.Pending
	}
	return 0
}

func (rjs *RunaiJobStatus) IsRunning() bool {
	if rjs.RunaiJob == nil || !rjs.Status.CompletionTime.IsZero() {
		return false
	}
	return rjs.GetReadyPods() > 0
}

func (rjs *RunaiJobStatus) IsCreating() bool {
	if rjs.RunaiJob == nil {
		// if there is no RunAI job, so workload is in creating state
		return true
	}
	// in case the workload completed (successfully or failed) or its suspended the job is not creating
	if !rjs.GetCompletionTime().IsZero() || rjs.IsSuspended() {
		return false
	}
	if rjs.RunaiJob.Status.Pending != 0 &&
		rjs.RunaiJob.Status.Initialized != 0 &&
		rjs.RunaiJob.Status.Running != 0 {
		return false
	}
	return true
}

func (rjs *RunaiJobStatus) IsInitializing() bool {
	if rjs.RunaiJob == nil || !rjs.GetCompletionTime().IsZero() {
		return false
	}
	if rjs.RunaiJob.Status.Ready < constants.MinimumReadyReplicas && rjs.RunaiJob.Status.Initialized > 0 {
		return true
	}
	return false
}

type RunaiJobAdapterFactory struct {
}

func (rjs *RunaiJobAdapterFactory) CreateJobAdapter(job client.Object) (JobAdapter, error) {
	runaiJob, ok := job.(*runaiJobV1.RunaiJob)
	if !ok {
		return nil, fmt.Errorf("job is not a runai job")
	}
	return NewRunaiJobAdapter(runaiJob), nil
}
