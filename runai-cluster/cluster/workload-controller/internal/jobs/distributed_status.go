/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"context"
	podPhase "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/pod"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type DistributedStatus struct {
	originalMetaObj             metav1.Object
	originalStatusObj           kubeflowcommon.JobStatus
	failedCondition             *kubeflowcommon.JobCondition
	succeededCondition          *kubeflowcommon.JobCondition
	suspendedCondition          *kubeflowcommon.JobCondition
	kubeClient                  client.Client
	numberOfExpectedPods        int32
	runningPods                 int32
	readyPods                   int32
	succeededPods               int32
	failedPods                  int32
	pendingPods                 int32
	initPods                    int32
	pods                        []podPhase.IRunaiPodPhaseExtended
	podPhaseCalculator          podPhase.IRunaiPhaseCalculator
	getManualPodLabelSelectorFn func(name string, replicaType kubeflowcommon.ReplicaType) *metav1.LabelSelector
}

func NewDistributedStatusFromJobStatus(jobStatus kubeflowcommon.JobStatus, metaObj metav1.Object, kubeClient client.Client,
	podPhaseCalculator podPhase.IRunaiPhaseCalculator, getManualPodLabelSelectorFn func(name string, replicaType kubeflowcommon.ReplicaType) *metav1.LabelSelector) *DistributedStatus {
	return &DistributedStatus{
		originalMetaObj:             metaObj,
		originalStatusObj:           jobStatus,
		kubeClient:                  kubeClient,
		podPhaseCalculator:          podPhaseCalculator,
		numberOfExpectedPods:        0,
		getManualPodLabelSelectorFn: getManualPodLabelSelectorFn,
	}
}

func (d *DistributedStatus) IsJobFinished() bool {
	return d.originalMetaObj != nil && !d.originalStatusObj.CompletionTime.IsZero()
}

func (d *DistributedStatus) IsTerminating() bool {
	return d.originalMetaObj != nil && !d.originalMetaObj.GetDeletionTimestamp().IsZero()
}

func (d *DistributedStatus) IsJobFailed() bool {
	if d.originalMetaObj == nil {
		return false
	}

	if d.failedCondition == nil {
		d.failedCondition = getCondition(d.originalStatusObj.Conditions, kubeflowcommon.JobFailed)
	}
	if d.failedCondition == nil {
		return false
	}
	return d.failedCondition.Status == corev1.ConditionTrue
}
func (d *DistributedStatus) IsJobSucceeded() bool {
	if d.originalMetaObj == nil {
		return false
	}

	if d.succeededCondition == nil {
		d.succeededCondition = getCondition(d.originalStatusObj.Conditions, kubeflowcommon.JobSucceeded)
	}
	if d.succeededCondition == nil {
		return false
	}
	return d.succeededCondition.Status == corev1.ConditionTrue
}

func (d *DistributedStatus) getDistributedPods(ctx context.Context) ([]podPhase.IRunaiPodPhaseExtended, error) {
	var pods = make([]podPhase.IRunaiPodPhaseExtended, 0)
	for replicaType, replica := range d.originalStatusObj.ReplicaStatuses {
		podList := &corev1.PodList{}
		podMatchingLabels, err := d.getLabelSelectorForReplicaPods(replicaType, replica)
		if err != nil {
			return nil, err
		}
		err = d.kubeClient.List(ctx, podList, podMatchingLabels, client.InNamespace(d.originalMetaObj.GetNamespace()))
		if err != nil {
			return nil, err
		}
		for _, pod := range podList.Items {
			// if the job does not own pod then skip it
			if !resources.IsObjectOwnedBy(&pod, d.originalMetaObj.GetUID()) {
				continue
			}
			pods = append(pods, d.podPhaseCalculator.Calculate(&pod))
		}
	}
	return pods, nil
}

func (d *DistributedStatus) getLabelSelectorForReplicaPods(replicaType kubeflowcommon.ReplicaType, replica *kubeflowcommon.ReplicaStatus) (*client.MatchingLabels, error) {
	podMatchingLabels := client.MatchingLabels{}
	labelSelector, err := getReplicaLabelSelector(replica)
	if err != nil {
		return nil, err
	}
	if labelSelector == nil {
		// manually create label selector
		labelSelector = d.getManualPodLabelSelectorFn(d.originalMetaObj.GetName(), replicaType)
	}
	for label, value := range labelSelector.MatchLabels {
		podMatchingLabels[label] = strings.ToLower(value)
	}
	return &podMatchingLabels, nil
}

func getReplicaLabelSelector(replica *kubeflowcommon.ReplicaStatus) (*metav1.LabelSelector, error) {
	if replica.Selector == "" {
		return replica.LabelSelector, nil
	}
	labelSelector, err := metav1.ParseToLabelSelector(replica.Selector)
	if err != nil {
		return nil, err
	}
	return labelSelector, nil
}

func (d *DistributedStatus) GetCompletionTime() *metav1.Time {
	return d.originalStatusObj.CompletionTime
}

// GetReadyPods - get the number of ready pods, which are pods with ready condition
func (d *DistributedStatus) GetReadyPods() int32 {
	if d.readyPods == 0 {
		if d.pods == nil {
			var err error
			d.pods, err = d.getDistributedPods(context.Background())
			if err != nil {
				return 0
			}
		}
		for _, pod := range d.pods {
			// if runai pod phase is created then are pending
			if pod.GetPhase() == podPhase.RunaiPhaseReady {
				d.readyPods++
			}
		}
	}
	return d.readyPods
}

func (d *DistributedStatus) GetRunningPods() int32 {
	if d.runningPods == 0 {
		if d.pods == nil {
			var err error
			d.pods, err = d.getDistributedPods(context.Background())
			if err != nil {
				return 0
			}
		}
		for _, pod := range d.pods {
			// if runai pod phase is created then are pending
			if pod.IsPodActive() {
				d.runningPods++
			}
		}
	}
	return d.runningPods
}

func (d *DistributedStatus) GetSucceededPods() int32 {
	if d.succeededPods == 0 {
		for _, replica := range d.originalStatusObj.ReplicaStatuses {
			d.succeededPods += replica.Succeeded
		}
	}
	return d.succeededPods
}

func (d *DistributedStatus) GetFailedPods() int32 {
	if d.failedPods == 0 {
		for _, replica := range d.originalStatusObj.ReplicaStatuses {
			d.failedPods += replica.Failed
		}
	}
	return d.failedPods
}

func (d *DistributedStatus) GetPendingPods() int32 {
	if d.pendingPods == 0 {
		if d.pods == nil {
			var err error
			d.pods, err = d.getDistributedPods(context.Background())
			if err != nil {
				return 0
			}
		}
		for _, pod := range d.pods {
			// if runai pod phase is created then are pending
			if pod.GetPhase() == podPhase.RunaiPhasePending {
				d.pendingPods++
			}
		}
	}
	return d.pendingPods
}

func (d *DistributedStatus) IsPending() bool {
	return d.originalMetaObj != nil && d.GetPendingPods() > 0
}

func (d *DistributedStatus) IsRunning() bool {
	return d.originalMetaObj != nil && d.GetReadyPods() > 0
}

func (d *DistributedStatus) IsInitializing() bool {
	return d.originalMetaObj != nil && d.GetInitializingPods() > 0 && d.GetReadyPods() < 1
}

func (d *DistributedStatus) IsCreating() bool {
	if d.originalMetaObj == nil ||
		d.originalStatusObj.ReplicaStatuses == nil ||
		len(d.originalStatusObj.ReplicaStatuses) == 0 {
		if !d.IsInitializing() && !d.IsRunning() && !d.IsPending() && !d.IsJobFinished() {
			return true
		}
	}
	return false
}

func (d *DistributedStatus) IsSuspended() bool {
	if d.originalMetaObj == nil {
		return false
	}

	if d.suspendedCondition == nil {
		d.suspendedCondition = getCondition(d.originalStatusObj.Conditions, kubeflowcommon.JobSuspended)
	}

	if d.suspendedCondition == nil {
		d.suspendedCondition = &kubeflowcommon.JobCondition{
			Type:   kubeflowcommon.JobSuspended,
			Status: corev1.ConditionFalse,
		}
		return false
	} else {
		return d.suspendedCondition.Status == corev1.ConditionTrue
	}
}

func (d *DistributedStatus) GetInitializingPods() int32 {
	if d.initPods == 0 {
		if d.pods == nil {
			var err error
			d.pods, err = d.getDistributedPods(context.Background())
			if err != nil {
				return 0
			}
		}
		for _, pod := range d.pods {
			if pod.GetPhase() == podPhase.RunaiPhaseInitializing {
				d.initPods++
			}
		}
	}
	return d.initPods
}

func (d *DistributedStatus) GetSuspendInfo() string {
	if d.IsSuspended() {
		if d.suspendedCondition != nil {
			return d.suspendedCondition.Message
		}
	}

	return constants.EmptyStr
}

func (d *DistributedStatus) GetFailedInfo() string {
	if d.failedCondition == nil {
		d.failedCondition = getCondition(d.originalStatusObj.Conditions, kubeflowcommon.JobFailed)
		if d.failedCondition == nil {
			return constants.EmptyStr
		}
	}
	var failedInfo string
	if d.failedCondition.Status == corev1.ConditionTrue {
		failedInfo = d.failedCondition.Message
	}
	return failedInfo
}

func (d *DistributedStatus) GetAllocatedGPU() string {
	if allocatedGpu, ok := d.originalMetaObj.GetAnnotations()[common.RunaiCurrentAllocatedGpusAnnotation]; ok {
		return allocatedGpu
	}
	return constants.EmptyStr
}

func (d *DistributedStatus) GetAllocatedGPUMemory() int64 {
	if memoryString, ok := d.originalMetaObj.GetAnnotations()[common.RunaiCurrentAllocatedGpuMemoryAnnotation]; ok {
		memory, err := strconv.ParseInt(memoryString, 10, 64)
		if err == nil {
			return memory
		}
	}
	return 0
}

func NewDistributedStatus() JobStatus {
	return &DistributedStatus{}
}

func getCondition(conditions []kubeflowcommon.JobCondition, conditionType kubeflowcommon.JobConditionType) *kubeflowcommon.JobCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func getMpiManualPodLabelSelector(name string, replicaType kubeflowcommon.ReplicaType) *metav1.LabelSelector {
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kubeflowcommon.JobNameLabel: name,
			kubeflowcommon.JobRoleLabel: string(replicaType),
		},
	}
	return labelSelector
}

func getTrainingJobManualPodLabelSelector(name string, replicaType kubeflowcommon.ReplicaType) *metav1.LabelSelector {
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			kubeflowcommon.JobNameLabel:     name,
			kubeflowcommon.ReplicaTypeLabel: string(replicaType),
		},
	}
	return labelSelector
}
