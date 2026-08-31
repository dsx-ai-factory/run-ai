/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"fmt"
	"sort"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	podphase "github.com/run-ai/runai/runai-cluster/cluster/common/utils/pod"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podPhaseCalc = &podphase.RunaiPhaseCalculator{}

type CriticalIssue struct {
	Group   string
	Compute bool
}

type Input struct {
	JobStatus *kubeflowv1.JobStatus
	Pods      []corev1.Pod
	Suspend   bool

	Terminating         bool
	CriticalIssues      []CriticalIssue
	PriorCompletionTime *metav1.Time
	MultiReplica        bool
}

type Result struct {
	Phase   statuscontract.Phase
	Message string

	CompletionTime *metav1.Time
	StartTime      *metav1.Time

	ReadyPods        int32
	RunningPods      int32
	PendingPods      int32
	InitializingPods int32
	SucceededPods    int32
	FailedPods       int32
}

func Calculate(in Input) Result {
	if in.JobStatus == nil {
		return Result{Phase: statuscontract.Creating}
	}

	ready, running, pending, initializing := classifyPods(in.Pods)
	succeeded, failed := replicaCounts(in.JobStatus)

	phase, msg := computePhase(in, ready, pending, initializing)

	completionTime := in.PriorCompletionTime
	if in.JobStatus.CompletionTime != nil {
		completionTime = in.JobStatus.CompletionTime
	}

	return Result{
		Phase:            phase,
		Message:          msg,
		CompletionTime:   completionTime,
		StartTime:        in.JobStatus.StartTime,
		ReadyPods:        ready,
		RunningPods:      running,
		PendingPods:      pending,
		InitializingPods: initializing,
		SucceededPods:    succeeded,
		FailedPods:       failed,
	}
}

func computePhase(in Input, readyPods, pendingPods, initPods int32) (statuscontract.Phase, string) {
	hasCriticalCompute, hasCritical, criticalGroups := classifyIssues(in.CriticalIssues)

	isRunning := readyPods > 0
	isJobFailed := conditionTrue(in.JobStatus, kubeflowv1.JobFailed)
	isSuspended := conditionTrue(in.JobStatus, kubeflowv1.JobSuspended) && readyPods == 0
	isJobFinished := in.JobStatus.CompletionTime != nil

	switch {
	case in.Terminating:
		return statuscontract.Terminating, ""
	case isJobFailed || hasCriticalCompute || (hasCritical && !isRunning):
		return statuscontract.Failed, failedMsg(in.JobStatus, isJobFailed, criticalGroups)
	case isJobFinished && !isJobFailed:
		return statuscontract.Completed, ""
	case in.Suspend && !isSuspended:
		return statuscontract.Suspending, ""
	case !in.Suspend && isSuspended:
		return statuscontract.Resuming, ""
	case isSuspended:
		return statuscontract.Suspended, conditionMessage(in.JobStatus, kubeflowv1.JobSuspended)
	case isRunning && hasCritical:
		return statuscontract.Degraded, degradedMsg(criticalGroups)
	case isRunning:
		return statuscontract.Running, ""
	case initPods > 0 && readyPods == 0:
		return statuscontract.Initializing, initializingMsg(in.MultiReplica, initPods)
	case pendingPods > 0:
		return statuscontract.Pending, ""
	default:
		return statuscontract.Creating, ""
	}
}

func classifyPods(pods []corev1.Pod) (ready, running, pending, initializing int32) {
	for i := range pods {
		ph := podPhaseCalc.Calculate(&pods[i]).GetPhase()
		switch ph {
		case podphase.RunaiPhaseReady:
			ready++
		case podphase.RunaiPhaseInitializing:
			initializing++
		case podphase.RunaiPhasePending:
			pending++
		}
		if pods[i].Status.Phase == corev1.PodRunning {
			running++
		}
	}
	return
}

func replicaCounts(js *kubeflowv1.JobStatus) (succeeded, failed int32) {
	for _, rs := range js.ReplicaStatuses {
		succeeded += rs.Succeeded
		failed += rs.Failed
	}
	return
}

func classifyIssues(issues []CriticalIssue) (hasCriticalCompute, hasCritical bool, groups []string) {
	groupSet := map[string]struct{}{}
	for _, iss := range issues {
		hasCritical = true
		groupSet[iss.Group] = struct{}{}
		if iss.Compute {
			hasCriticalCompute = true
		}
	}
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return
}

func conditionTrue(js *kubeflowv1.JobStatus, ct kubeflowv1.JobConditionType) bool {
	for _, c := range js.Conditions {
		if c.Type == ct {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func conditionMessage(js *kubeflowv1.JobStatus, ct kubeflowv1.JobConditionType) string {
	for _, c := range js.Conditions {
		if c.Type == ct {
			return c.Message
		}
	}
	return ""
}

func initializingMsg(multiReplica bool, initPods int32) string {
	if multiReplica {
		return fmt.Sprintf(statuscontract.MsgInitializingMultiPods, initPods)
	}
	return statuscontract.MsgInitializingSinglePod
}

func failedMsg(js *kubeflowv1.JobStatus, isJobFailed bool, criticalGroups []string) string {
	if isJobFailed {
		if msg := conditionMessage(js, kubeflowv1.JobFailed); msg != "" {
			return msg
		}
		return statuscontract.MsgGenericJobFailed
	}
	return statuscontract.FormatIssueCause(criticalGroups,
		statuscontract.MsgFailedDueToIssueSingular, statuscontract.MsgFailedDueToIssuesPlural)
}

func degradedMsg(criticalGroups []string) string {
	return statuscontract.FormatIssueCause(criticalGroups,
		statuscontract.MsgDegradedDueToIssueSingular, statuscontract.MsgDegradedDueToIssuesPlural)
}
