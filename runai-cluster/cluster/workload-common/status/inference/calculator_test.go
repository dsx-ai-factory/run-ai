/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"testing"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCalculate(t *testing.T) {
	readyPod := pod(corev1.PodRunning, true, "", "")
	pendingPod := pod(corev1.PodPending, false, "", "")
	containerCreatingPod := pod(corev1.PodPending, false, "ContainerCreating", "")
	imagePullPod := pod(corev1.PodPending, false, "", "ImagePullBackOff")

	tests := []struct {
		name    string
		input   Input
		phase   statuscontract.Phase
		message string
	}{
		{
			name:  "missing service is creating",
			phase: statuscontract.Creating,
		},
		{
			name:  "deletion takes precedence",
			input: Input{MarkedForDeletion: true},
			phase: statuscontract.Deleting,
		},
		{
			name:  "service without a created revision is creating",
			input: Input{Service: &ServiceState{}},
			phase: statuscontract.Creating,
		},
		{
			name:  "created revision without pods is pending",
			input: Input{Service: &ServiceState{LatestCreatedRevision: "rev-1"}},
			phase: statuscontract.Pending,
		},
		{
			name: "pending pod is pending",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1"},
				Pods:    []corev1.Pod{pendingPod},
			},
			phase: statuscontract.Pending,
		},
		{
			name: "container creation is initializing",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1"},
				Pods:    []corev1.Pod{containerCreatingPod},
			},
			phase:   statuscontract.Initializing,
			message: "Workload is pulling container images.",
		},
		{
			name: "image pull retry is initializing",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1"},
				Pods:    []corev1.Pod{imagePullPod},
			},
			phase:   statuscontract.Initializing,
			message: "Failed to pull container image. Retrying...",
		},
		{
			name: "running pod that is not ready is initializing",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1"},
				Pods:    []corev1.Pod{pod(corev1.PodRunning, false, "", "")},
			},
			phase:   statuscontract.Initializing,
			message: "Container images pulled successfully. Workload is initializing.",
		},
		{
			name: "ready revision with ready pod is ready",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1", LatestReadyRevision: "rev-1"},
				Pods:    []corev1.Pod{readyPod},
			},
			phase: statuscontract.Ready,
		},
		{
			name: "ready revision without pods supports scale to zero",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1", LatestReadyRevision: "rev-1"},
			},
			phase: statuscontract.Ready,
		},
		{
			name: "ready revision with pending pod is pending after scale up",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1", LatestReadyRevision: "rev-1"},
				Pods:    []corev1.Pod{pendingPod},
			},
			phase: statuscontract.Pending,
		},
		{
			name: "older ready revision keeps an update ready",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-2", LatestReadyRevision: "rev-1"},
				Pods:    []corev1.Pod{pendingPod},
			},
			phase:   statuscontract.Ready,
			message: "Workload is ready. Update in progress: waiting for resources.",
		},
		{
			name:    "image pull condition fails an initial rollout",
			input:   Input{Service: failedService("rev-1", "", "unable to fetch image")},
			phase:   statuscontract.Failed,
			message: "Failed to pull container image.",
		},
		{
			name:    "name length condition reports actionable failure",
			input:   Input{Service: failedService("rev-1", "", "name part must be no more than 63 characters")},
			phase:   statuscontract.Failed,
			message: "The workload name is too long and causes derived resource names to exceed Kubernetes 63 character limit. Please use a shorter workload name.",
		},
		{
			name:    "unknown false condition reports timeout",
			input:   Input{Service: failedService("rev-1", "", "revision failed")},
			phase:   statuscontract.Failed,
			message: "Timed out waiting for the workload to become ready.",
		},
		{
			name:    "failure on current ready revision is degraded",
			input:   Input{Service: failedService("rev-1", "rev-1", "unable to fetch image")},
			phase:   statuscontract.Degraded,
			message: "Failed to pull container image.",
		},
		{
			name:    "failed update keeps the older revision ready",
			input:   Input{Service: failedService("rev-2", "rev-1", "unable to fetch image")},
			phase:   statuscontract.Ready,
			message: "Workload is ready. Update failed: Failed to pull container image.",
		},
		{
			name: "build error fails initial rollout",
			input: Input{
				Service:    &ServiceState{LatestCreatedRevision: "rev-1"},
				BuildError: stringPointer("admission denied"),
			},
			phase:   statuscontract.Failed,
			message: "Failed to create workload: admission denied",
		},
		{
			name: "knative prerequisite build error is enriched",
			input: Input{
				Service: &ServiceState{LatestCreatedRevision: "rev-1"},
				BuildError: stringPointer("validation.webhook.serving.knative.dev denied the request: " +
					"Persistent volume claim support is disabled"),
			},
			phase: statuscontract.Failed,
			message: "Failed to create workload: validation.webhook.serving.knative.dev denied the request: " +
				"Persistent volume claim support is disabled. Check the inference prerequisites in the Run:ai documentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Calculate(tt.input)
			if result.Phase != tt.phase || result.Message != tt.message {
				t.Fatalf("Calculate() = (%q, %q), want (%q, %q)",
					result.Phase, result.Message, tt.phase, tt.message)
			}
		})
	}
}

func TestCalculateMixedPendingAndImagePullIsOrderIndependent(t *testing.T) {
	pending := pod(corev1.PodPending, false, "", "")
	imagePull := pod(corev1.PodPending, false, "", "ImagePullBackOff")
	service := &ServiceState{LatestCreatedRevision: "rev-1"}

	for _, pods := range [][]corev1.Pod{{pending, imagePull}, {imagePull, pending}} {
		result := Calculate(Input{Service: service, Pods: pods})
		if result.Phase != statuscontract.Pending || result.Message != "" {
			t.Fatalf("Calculate() = (%q, %q), want Pending with empty message", result.Phase, result.Message)
		}
	}
}

func TestCalculateRecognizesAllImagePullReasons(t *testing.T) {
	service := &ServiceState{LatestCreatedRevision: "rev-1"}
	for _, reason := range []string{"ImagePullBackOff", "ErrImagePull", "ErrImageNeverPull"} {
		t.Run(reason, func(t *testing.T) {
			result := Calculate(Input{
				Service: service,
				Pods:    []corev1.Pod{pod(corev1.PodPending, false, "", reason)},
			})
			if result.Phase != statuscontract.Initializing || result.Message != "Failed to pull container image. Retrying..." {
				t.Fatalf("Calculate() = (%q, %q), want Initializing with image-pull retry message",
					result.Phase, result.Message)
			}
		})
	}
}

func TestCalculateReportsCurrentRevisionCrashLoop(t *testing.T) {
	const timeout = "Timed out waiting for the workload to become ready."
	exit0 := crashLoopPod("rev-1", &corev1.ContainerStateTerminated{ExitCode: 0, Reason: "Completed"})
	exit1 := crashLoopPod("rev-1", &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"})
	slowStart := pod(corev1.PodRunning, false, "", "")
	slowStart.Labels = map[string]string{"serving.knative.dev/revision": "rev-1"}
	slowStart.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name:  "user-container",
		State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
	}}
	unlabeled := crashLoopPod("rev-1", &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"})
	unlabeled.Labels = nil
	restartedRunning := crashLoopPod("rev-1", &corev1.ContainerStateTerminated{ExitCode: 0, Reason: "Completed"})
	restartedRunning.Status.ContainerStatuses[0].State = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}

	tests := []struct {
		name    string
		input   Input
		phase   statuscontract.Phase
		message string
	}{
		{
			name:    "crash loop exiting with code 0 reports the container instead of a timeout",
			input:   Input{Service: failedService("rev-1", "", "revision failed"), Pods: []corev1.Pod{exit0}},
			phase:   statuscontract.Failed,
			message: `Container "user-container" keeps crashing: restarted 4 times, last exit code 0 (Completed).`,
		},
		{
			name:    "crash loop exiting with code 1 on the ready revision is degraded with crash details",
			input:   Input{Service: failedService("rev-1", "rev-1", "revision failed"), Pods: []corev1.Pod{exit1}},
			phase:   statuscontract.Degraded,
			message: `Container "user-container" keeps crashing: restarted 4 times, last exit code 1 (Error).`,
		},
		{
			name:    "crash loop without a recorded termination reports restarts only",
			input:   Input{Service: failedService("rev-1", "", "revision failed"), Pods: []corev1.Pod{crashLoopPod("rev-1", nil)}},
			phase:   statuscontract.Failed,
			message: `Container "user-container" keeps crashing: restarted 4 times.`,
		},
		{
			name: "crash loop in a failed update keeps the update prefix",
			input: Input{
				Service: failedService("rev-2", "rev-1", "revision failed"),
				Pods:    []corev1.Pod{crashLoopPod("rev-2", &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"})},
			},
			phase:   statuscontract.Ready,
			message: `Workload is ready. Update failed: Container "user-container" keeps crashing: restarted 4 times, last exit code 1 (Error).`,
		},
		{
			name:    "crashed container running again between back-offs still reports crash details",
			input:   Input{Service: failedService("rev-1", "", "revision failed"), Pods: []corev1.Pod{restartedRunning}},
			phase:   statuscontract.Failed,
			message: `Container "user-container" keeps crashing: restarted 4 times, last exit code 0 (Completed).`,
		},
		{
			name:    "slow start without restarts still reports a timeout",
			input:   Input{Service: failedService("rev-1", "", "revision failed"), Pods: []corev1.Pod{slowStart}},
			phase:   statuscontract.Failed,
			message: timeout,
		},
		{
			name: "crash loop from an older revision is ignored",
			input: Input{
				Service: failedService("rev-2", "", "revision failed"),
				Pods:    []corev1.Pod{crashLoopPod("rev-1", &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"})},
			},
			phase:   statuscontract.Failed,
			message: timeout,
		},
		{
			name:    "crash loop on a pod without a revision label is ignored",
			input:   Input{Service: failedService("rev-1", "", "revision failed"), Pods: []corev1.Pod{unlabeled}},
			phase:   statuscontract.Failed,
			message: timeout,
		},
		{
			name:    "image pull failure keeps priority over crash details",
			input:   Input{Service: failedService("rev-1", "", "unable to fetch image"), Pods: []corev1.Pod{exit1}},
			phase:   statuscontract.Failed,
			message: "Failed to pull container image.",
		},
		{
			name:    "name length failure keeps priority over crash details",
			input:   Input{Service: failedService("rev-1", "", "must be no more than 63 characters"), Pods: []corev1.Pod{exit1}},
			phase:   statuscontract.Failed,
			message: "The workload name is too long and causes derived resource names to exceed Kubernetes 63 character limit. Please use a shorter workload name.",
		},
		{
			name: "build error keeps priority over crash details",
			input: Input{
				Service:    failedService("rev-1", "", "revision failed"),
				Pods:       []corev1.Pod{exit1},
				BuildError: stringPointer("admission denied"),
			},
			phase:   statuscontract.Failed,
			message: "Failed to create workload: admission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Calculate(tt.input)
			if result.Phase != tt.phase || result.Message != tt.message {
				t.Fatalf("Calculate() = (%q, %q), want (%q, %q)",
					result.Phase, result.Message, tt.phase, tt.message)
			}
		})
	}
}

func crashLoopPod(revision string, lastTermination *corev1.ContainerStateTerminated) corev1.Pod {
	pod := pod(corev1.PodRunning, false, "", "")
	pod.Labels = map[string]string{"serving.knative.dev/revision": revision}
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name:                 "user-container",
		RestartCount:         4,
		State:                corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
		LastTerminationState: corev1.ContainerState{Terminated: lastTermination},
	}}
	return pod
}

func failedService(created, ready, message string) *ServiceState {
	return &ServiceState{
		LatestCreatedRevision: created,
		LatestReadyRevision:   ready,
		Conditions: []ServiceCondition{{
			Status:  corev1.ConditionFalse,
			Message: message,
		}},
	}
}

func pod(phase corev1.PodPhase, ready bool, calculatedPhase, waitingReason string) corev1.Pod {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
		Status:     corev1.PodStatus{Phase: phase},
	}
	if calculatedPhase != "" {
		pod.Annotations[commonconstants.PodStatusAnnotation] = calculatedPhase
	}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	if waitingReason != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
			State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: waitingReason}},
		}}
	}
	return pod
}

func stringPointer(value string) *string {
	return &value
}
