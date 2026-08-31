/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	"fmt"
	"sort"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// RuntimeIssueGroup is the group for a runtime pod failure (always compute).
const RuntimeIssueGroup = "Compute"

// Kubelet container-state reasons (not exposed as typed constants in k8s.io/api).
const (
	reasonOOMKilled                = "OOMKilled"
	reasonRunContainerError        = "RunContainerError"
	reasonContainerCannotRun       = "ContainerCannotRun"
	reasonCreateContainerError     = "CreateContainerError"
	reasonCreateContainerConfigErr = "CreateContainerConfigError"
	reasonInvalidImageName         = "InvalidImageName"
	reasonErrImagePull             = "ErrImagePull"
	reasonImagePullBackOff         = "ImagePullBackOff"
)

// ComputeRuntimeIssue returns the group and pod-level cause when the workload failed and the
// wlr surfaced no critical issue.
func ComputeRuntimeIssue(job *batchv1.Job, pods []corev1.Pod, wlr *dromosv1alpha1.WorkloadRunner) (group, desc string, ok bool) {
	failed := job != nil && (jobConditionTrue(job, batchv1.JobFailed) || wlrPhase(wlr) == dromosv1alpha1.WorkloadRunnerPhaseFailed)
	return ComputeRuntimeIssueFromPods(failed, pods, wlr)
}

func ComputeRuntimeIssueFromPods(failed bool, pods []corev1.Pod, wlr *dromosv1alpha1.WorkloadRunner) (group, desc string, ok bool) {
	if !failed || wlrHasCriticalIssue(wlr) {
		return "", "", false
	}
	ordered := append([]corev1.Pod(nil), pods...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	for i := range ordered {
		if msg := containerFailureMessage(&ordered[i]); msg != "" {
			if name := ordered[i].Name; name != "" {
				return RuntimeIssueGroup, fmt.Sprintf("%s (pod: %s)", msg, name), true
			}
			return RuntimeIssueGroup, msg, true
		}
	}
	return RuntimeIssueGroup, "", true
}

func wlrHasCriticalIssue(wlr *dromosv1alpha1.WorkloadRunner) bool {
	if wlr == nil {
		return false
	}
	for _, is := range wlr.Status.Issues {
		if is.IsCritical {
			return true
		}
	}
	for _, plugin := range wlr.Status.Plugins {
		if plugin == nil {
			continue
		}
		for _, is := range plugin.Issues {
			if is.IsCritical {
				return true
			}
		}
	}
	return false
}

// containerFailureMessage returns the first failure message, init containers before app ones.
func containerFailureMessage(pod *corev1.Pod) string {
	for _, cs := range pod.Status.InitContainerStatuses {
		if msg := containerStatusMessage(cs); msg != "" {
			return msg
		}
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if msg := containerStatusMessage(cs); msg != "" {
			return msg
		}
	}
	return ""
}

func containerStatusMessage(cs corev1.ContainerStatus) string {
	if msg := terminatedMessage(cs.Name, cs.State.Terminated); msg != "" {
		return msg
	}
	return waitingMessage(cs.Name, cs.State.Waiting)
}

func terminatedMessage(name string, t *corev1.ContainerStateTerminated) string {
	if t == nil || t.ExitCode == 0 {
		return ""
	}
	// OOMKilled gets a named message; the exit-code fallback (137) is too generic.
	if t.Reason == reasonOOMKilled {
		return fmt.Sprintf("container '%s' was OOMKilled", name)
	}
	if t.Message != "" {
		return fmt.Sprintf("container '%s' failed: %s", name, t.Message)
	}
	return fmt.Sprintf("container '%s' exited with exit code %d", name, t.ExitCode)
}

// waitingMessage surfaces errors from containers that failed to start (e.g. RunContainerError).
func waitingMessage(name string, w *corev1.ContainerStateWaiting) string {
	if w == nil {
		return ""
	}
	switch w.Reason {
	case reasonRunContainerError, reasonContainerCannotRun, reasonCreateContainerError,
		reasonCreateContainerConfigErr, reasonInvalidImageName,
		reasonErrImagePull, reasonImagePullBackOff:
		if w.Message != "" {
			return fmt.Sprintf("container '%s' failed to start: %s", name, w.Message)
		}
		return fmt.Sprintf("container '%s' failed to start (%s)", name, w.Reason)
	}
	return ""
}
