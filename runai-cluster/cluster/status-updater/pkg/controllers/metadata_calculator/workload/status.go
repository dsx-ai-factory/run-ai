/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	podstatusutils "github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/pod"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	runaiJobKind    = "RunaiJob"
	jobKind         = "Job"
	deploymentKind  = "Deployment"
	statefulSetKind = "StatefulSet"
	replicaSetKind  = "ReplicaSet"

	pendingStatus   = "Pending"
	runningStatus   = "Running"
	succeededStatus = "Succeeded"
	failedStatus    = "Failed"
	deletedStatus   = "Deleted"
	preemptedStatus = "Preempted"
	timeoutStatus   = "Timeout"
	suspendedStatus = "Suspended"

	deadlineExceeded = "DeadlineExceeded"

	trueStr = "True"
)

func CalcWorkloadStatus(ctx context.Context, workloadRelatedObjects *metadata.WorkloadObjects,
	kubeClient client.Client, workloadMetadata *metadata.WorkloadMetadata) error {
	var err error
	workloadMetadata.Status, err = getWorkloadStatus(ctx, workloadRelatedObjects, kubeClient)
	return err
}

//nolint:gocognit
func getWorkloadStatus(ctx context.Context, workloadRelatedObjects *metadata.WorkloadObjects,
	kubeClient client.Client) (string, error) {
	var err error
	var runaiOwnerJob *runaiv1.RunaiJob
	var ownerJob *batchv1.Job

	ownerObject := workloadRelatedObjects.WorkloadObj
	podsExtendedPhases := map[string]int{}
	for _, item := range workloadRelatedObjects.Pods {
		pod := item
		podsExtendedPhases[podstatusutils.GetExtendedPhase(&pod)]++
	}

	//nolint:nestif // unnecessary complex branching tree leads to unnecessary complex code
	if ownerObject.GetKind() == runaiJobKind {
		if runaiOwnerJob, err = getRunaiJob(ctx, ownerObject, workloadRelatedObjects, kubeClient); err != nil {
			return "", err
		}

		status := handleRunaiJobSpecificStatuses(runaiOwnerJob)
		if status != "" {
			return status, nil
		}
	} else if ownerObject.GetKind() == jobKind {
		if ownerJob, err = getJob(ctx, ownerObject, workloadRelatedObjects, kubeClient); err != nil {
			return "", err
		}

		status := handleJobSpecificStatuses(ownerJob)
		if status != "" {
			return status, nil
		}
	}

	if podsExtendedPhases[runningStatus] > 0 {
		return runningStatus, nil
	}

	//nolint:nestif // unnecessary complex branching tree leads to unnecessary complex code
	if ownerObject.GetKind() == runaiJobKind {
		status := handleRunaiJobFailure(runaiOwnerJob)
		if status != "" {
			return status, nil
		}

		if runaiOwnerJob.Spec.Parallelism != nil && *runaiOwnerJob.Spec.Parallelism > 1 { // handle HPO runai jobs
			return handleHpo(runaiOwnerJob.Spec.Completions, runaiOwnerJob.Status.Succeeded,
				workloadRelatedObjects.Pods, podsExtendedPhases)
		}
	} else if ownerObject.GetKind() == jobKind {
		status := handleJobFailure(ownerJob)
		if status != "" {
			return status, nil
		}

		if ownerJob.Spec.Parallelism != nil && *ownerJob.Spec.Parallelism > 1 { // handle HPO jobs
			return handleHpo(ownerJob.Spec.Completions, ownerJob.Status.Succeeded,
				workloadRelatedObjects.Pods, podsExtendedPhases)
		}
	}

	if workloadRelatedObjects.LatestChiefPod == nil {
		return deletedStatus, nil
	}

	chiefPodStatus := podstatusutils.TranslatePodStatus(workloadRelatedObjects.LatestChiefPod)
	chiefPodExtendedStatus := podstatusutils.GetExtendedPhase(workloadRelatedObjects.LatestChiefPod)

	if (ownerObject.GetKind() == deploymentKind ||
		ownerObject.GetKind() == statefulSetKind ||
		ownerObject.GetKind() == replicaSetKind) &&
		isFinishedStatus(chiefPodStatus) {
		if podstatusutils.IsPodBound(chiefPodExtendedStatus) {
			return runningStatus, nil
		} else {
			return pendingStatus, nil
		}
	}

	if chiefPodStatus == deletedStatus || chiefPodStatus == failedStatus {
		// Workload isn't finished yet, but  pods is, waiting for new one to be created
		return pendingStatus, nil
	}

	return chiefPodStatus, nil
}

func getRunaiJob(ctx context.Context, ownerObject *unstructured.Unstructured,
	workloadData *metadata.WorkloadObjects, kubeClient client.Client) (*runaiv1.RunaiJob, error) {
	runaiOwnerJob := &runaiv1.RunaiJob{}
	err := kubeClient.Get(ctx,
		types.NamespacedName{Namespace: workloadData.Namespace, Name: ownerObject.GetName()}, runaiOwnerJob)
	if err != nil {
		return nil, err
	}
	return runaiOwnerJob, nil
}

func getJob(ctx context.Context, ownerObj *unstructured.Unstructured,
	workloadData *metadata.WorkloadObjects, kubeClient client.Client) (*batchv1.Job, error) {
	ownerJob := &batchv1.Job{}
	err := kubeClient.Get(ctx, types.NamespacedName{Namespace: workloadData.Namespace, Name: ownerObj.GetName()}, ownerJob)
	if err != nil {
		return nil, err
	}
	return ownerJob, nil
}

func handleRunaiJobSpecificStatuses(runaiOwnerJob *runaiv1.RunaiJob) string {
	if runaiOwnerJob.DeletionTimestamp != nil {
		return deletedStatus
	}
	if runaiOwnerJob.Spec.Suspend != nil && *runaiOwnerJob.Spec.Suspend {
		return suspendedStatus
	}
	for _, event := range runaiOwnerJob.Status.Conditions {
		if event.Type == failedStatus && event.Status == trueStr {
			if event.Reason == deadlineExceeded {
				return timeoutStatus
			}
		}
	}
	return ""
}

func handleJobSpecificStatuses(ownerJob *batchv1.Job) string {
	if ownerJob.DeletionTimestamp != nil {
		return deletedStatus
	} else {
		for _, event := range ownerJob.Status.Conditions {
			if event.Type == failedStatus && event.Status == trueStr {
				if event.Reason == deadlineExceeded {
					return timeoutStatus
				}
			}
		}
	}
	return ""
}

func handleRunaiJobFailure(runaiOwnerJob *runaiv1.RunaiJob) string {
	for _, event := range runaiOwnerJob.Status.Conditions {
		if event.Type == failedStatus && event.Status == trueStr {
			return failedStatus
		}
	}
	return ""
}

func handleJobFailure(ownerJob *batchv1.Job) string {
	for _, event := range ownerJob.Status.Conditions {
		if event.Type == failedStatus && event.Status == trueStr {
			return failedStatus
		}
	}
	return ""
}

func handleHpo(jobCompletions *int32, jobSucceededCount int32, jobPods []v1.Pod, podExtendedPhases map[string]int) (
	string, error) {
	completions := int32(1)
	if jobCompletions != nil {
		completions = *jobCompletions
	}

	if jobSucceededCount >= completions || int32(podExtendedPhases[succeededStatus]) >= completions {
		return succeededStatus, nil
	}

	if len(jobPods) > 0 {
		if checkIfAllPodsHaveTheSameStatus(jobPods) {
			return podstatusutils.TranslatePodStatus(&jobPods[0]), nil
		}
	}

	if podExtendedPhases[pendingStatus] > 0 {
		return pendingStatus, nil
	}
	if podExtendedPhases[failedStatus] > 0 {
		return failedStatus, nil
	} else {
		return pendingStatus, nil
	}
}

func checkIfAllPodsHaveTheSameStatus(jobPods []v1.Pod) bool {
	sameStatus := ""
	for _, item := range jobPods {
		pod := item
		currentPodStatus := podstatusutils.TranslatePodStatus(&pod)
		if sameStatus != "" && currentPodStatus != sameStatus {
			sameStatus = ""
			break
		}
		sameStatus = currentPodStatus
	}
	return sameStatus != ""
}

func isFinishedStatus(toCheck string) bool {
	return toCheck == succeededStatus || toCheck == failedStatus || toCheck == deletedStatus ||
		toCheck == preemptedStatus || toCheck == timeoutStatus
}
