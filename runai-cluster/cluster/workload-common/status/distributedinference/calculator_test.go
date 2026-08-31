/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributedinference

import (
	"testing"
	"time"

	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestDeterminePodPhase(t *testing.T) {
	now := metav1.NewTime(time.Now())
	tests := []struct {
		name string
		pod  corev1.Pod
		want podPhase
	}{
		{name: "deleting", pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now}}, want: podPhaseFailed},
		{name: "running ready", pod: groupPod("0", "0", corev1.PodRunning, true, ""), want: podPhaseReady},
		{name: "running not ready", pod: groupPod("0", "0", corev1.PodRunning, false, ""), want: podPhaseRunning},
		{name: "succeeded", pod: groupPod("0", "0", corev1.PodSucceeded, false, ""), want: podPhaseRunning},
		{name: "failed", pod: groupPod("0", "0", corev1.PodFailed, false, ""), want: podPhaseFailed},
		{name: "container creating", pod: groupPod("0", "0", corev1.PodPending, false, "ContainerCreating"), want: podPhaseInitializing},
		{name: "pod initializing", pod: groupPod("0", "0", corev1.PodPending, false, "PodInitializing"), want: podPhaseInitializing},
		{name: "image pull backoff", pod: groupPod("0", "0", corev1.PodPending, false, "ImagePullBackOff"), want: podPhaseInitializing},
		{name: "image pull error", pod: groupPod("0", "0", corev1.PodPending, false, "ErrImagePull"), want: podPhaseInitializing},
		{name: "never pull", pod: groupPod("0", "0", corev1.PodPending, false, "ErrImageNeverPull"), want: podPhaseInitializing},
		{name: "container error", pod: groupPod("0", "0", corev1.PodPending, false, "CreateContainerError"), want: podPhaseFailed},
		{name: "container config error", pod: groupPod("0", "0", corev1.PodPending, false, "CreateContainerConfigError"), want: podPhaseFailed},
		{name: "other pending reason", pod: groupPod("0", "0", corev1.PodPending, false, "Unschedulable"), want: podPhasePending},
		{name: "unknown phase", pod: groupPod("0", "0", corev1.PodUnknown, false, ""), want: podPhasePending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determinePodPhase(&tt.pod); got != tt.want {
				t.Fatalf("determinePodPhase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetermineGroupPhase(t *testing.T) {
	tests := []struct {
		name         string
		group        map[string]podPhase
		expectedSize int
		want         statuscontract.Phase
		wantErr      bool
	}{
		{name: "incomplete", group: map[string]podPhase{"0": podPhaseReady}, expectedSize: 2, want: statuscontract.Pending},
		{name: "failed has priority", group: map[string]podPhase{"0": podPhaseReady, "1": podPhaseFailed}, expectedSize: 2, want: statuscontract.Failed},
		{name: "pending precedes initializing", group: map[string]podPhase{"0": podPhasePending, "1": podPhaseInitializing}, expectedSize: 2, want: statuscontract.Pending},
		{name: "initializing precedes readiness", group: map[string]podPhase{"0": podPhaseReady, "1": podPhaseInitializing}, expectedSize: 2, want: statuscontract.Initializing},
		{name: "ready leader", group: map[string]podPhase{"0": podPhaseReady, "1": podPhaseRunning}, expectedSize: 2, want: statuscontract.Ready},
		{name: "running leader", group: map[string]podPhase{"0": podPhaseRunning, "1": podPhaseReady}, expectedSize: 2, want: statuscontract.Initializing},
		{name: "missing leader", group: map[string]podPhase{"1": podPhaseReady}, expectedSize: 1, want: statuscontract.Failed, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineGroupPhase(tt.group, tt.expectedSize)
			if (err != nil) != tt.wantErr {
				t.Fatalf("determineGroupPhase() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("determineGroupPhase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculateRuntimeReportsMissingLeader(t *testing.T) {
	result, err := CalculateRuntime(RuntimeInput{
		Replicas: 1, GroupSize: 1,
		Pods: []corev1.Pod{groupPod("0", "1", corev1.PodRunning, true, "")},
	})

	if err == nil {
		t.Fatal("CalculateRuntime() error = nil, want missing leader error")
	}
	if result.Phase != statuscontract.Failed {
		t.Fatalf("CalculateRuntime() phase = %q, want %q", result.Phase, statuscontract.Failed)
	}
}

func TestAggregatePhase(t *testing.T) {
	tests := []struct {
		name string
		info phaseInfo
		want statuscontract.Phase
	}{
		{name: "zero replicas", info: phaseInfo{}, want: statuscontract.Ready},
		{name: "ready and failed", info: phaseInfo{replicas: 2, ready: 1, failed: 1}, want: statuscontract.Degraded},
		{name: "failed without ready", info: phaseInfo{replicas: 2, failed: 1, pending: 1}, want: statuscontract.Failed},
		{name: "all pending", info: phaseInfo{replicas: 2, pending: 2}, want: statuscontract.Pending},
		{name: "partly ready", info: phaseInfo{replicas: 2, ready: 1, pending: 1}, want: statuscontract.Initializing},
		{name: "all ready", info: phaseInfo{replicas: 2, ready: 2}, want: statuscontract.Ready},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aggregatePhase(tt.info); got != tt.want {
				t.Fatalf("aggregatePhase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name  string
		input Input
		want  Result
	}{
		{
			name: "missing LWS is creating",
			want: Result{Phase: statuscontract.Creating},
		},
		{
			name:  "missing LWS with build failure is failed",
			input: Input{BuildError: ptr.To("LeaderWorkerSet not found")},
			want:  Result{Phase: statuscontract.Failed},
		},
		{
			name: "deletion takes precedence",
			input: Input{
				Exists: true, Replicas: 1, GroupSize: 1, MarkedForDeletion: true,
				Pods: []corev1.Pod{groupPod("0", "0", corev1.PodRunning, true, "")},
			},
			want: Result{Phase: statuscontract.Deleting, Replicas: 1, ReadyReplicas: 1},
		},
		{
			name:  "zero replicas is ready",
			input: Input{Exists: true, Replicas: 0, GroupSize: 1},
			want:  Result{Phase: statuscontract.Ready, Message: "Replicas: 0"},
		},
		{
			name: "missing resources degrade ready workload",
			input: Input{Exists: true, Replicas: 0, GroupSize: 1,
				MissingResources: []string{"cpu", "gpu"}},
			want: Result{Phase: statuscontract.Degraded, Message: "Replicas: 0; Missing resources: cpu, gpu"},
		},
		{
			name: "build error takes precedence over missing resources",
			input: Input{Exists: true, Replicas: 0, GroupSize: 1,
				BuildError: ptr.To("build failed"), MissingResources: []string{"gpu"}},
			want: Result{Phase: statuscontract.Degraded, Message: "Replicas: 0; build failed"},
		},
		{
			name:  "empty build error does not add message punctuation",
			input: Input{Exists: true, Replicas: 0, GroupSize: 1, BuildError: ptr.To("")},
			want:  Result{Phase: statuscontract.Degraded, Message: "Replicas: 0"},
		},
		{
			name: "MNNVL not ready degrades ready workload",
			input: Input{Exists: true, Replicas: 0, GroupSize: 1,
				MNNVLEnabled: true},
			want: Result{Phase: statuscontract.Degraded,
				Message: "Replicas: 0; " + mnnvltypes.UserFacingFeatureName + " is not ready"},
		},
		{
			name: "one ready and one failed group is degraded",
			input: Input{Exists: true, Replicas: 2, GroupSize: 1, Pods: []corev1.Pod{
				groupPod("0", "0", corev1.PodRunning, true, ""),
				groupPod("1", "0", corev1.PodFailed, false, ""),
			}},
			want: Result{Phase: statuscontract.Degraded, Message: "Replicas: 2; Ready: 1; Failed: 1",
				Replicas: 2, ReadyReplicas: 1},
		},
		{
			name: "missing group is initializing beside ready group",
			input: Input{Exists: true, Replicas: 2, GroupSize: 1, Pods: []corev1.Pod{
				groupPod("0", "0", corev1.PodRunning, true, ""),
			}},
			want: Result{Phase: statuscontract.Initializing, Message: "Replicas: 2; Ready: 1; Pending: 1",
				Replicas: 2, ReadyReplicas: 1},
		},
		{
			name: "ready leader keeps group ready while worker is running",
			input: Input{Exists: true, Replicas: 1, GroupSize: 2, Pods: []corev1.Pod{
				groupPod("0", "0", corev1.PodRunning, true, ""),
				groupPod("0", "1", corev1.PodRunning, false, ""),
			}},
			want: Result{Phase: statuscontract.Ready, Message: "Replicas: 1; Ready: 1",
				Replicas: 1, ReadyReplicas: 1},
		},
		{
			name: "group without leader fails",
			input: Input{Exists: true, Replicas: 1, GroupSize: 1, Pods: []corev1.Pod{
				groupPod("0", "1", corev1.PodRunning, true, ""),
			}},
			want: Result{Phase: statuscontract.Failed, Message: "Replicas: 1; Failed: 1", Replicas: 1},
		},
		{
			name: "image pull initializes group",
			input: Input{Exists: true, Replicas: 1, GroupSize: 1, Pods: []corev1.Pod{
				groupPod("0", "0", corev1.PodPending, false, "ImagePullBackOff"),
			}},
			want: Result{Phase: statuscontract.Initializing, Message: "Replicas: 1; Initializing: 1", Replicas: 1},
		},
		{
			name: "container creation error fails group",
			input: Input{Exists: true, Replicas: 1, GroupSize: 1, Pods: []corev1.Pod{
				groupPod("0", "0", corev1.PodPending, false, "CreateContainerConfigError"),
			}},
			want: Result{Phase: statuscontract.Failed, Message: "Replicas: 1; Failed: 1", Replicas: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Calculate(tt.input); got != tt.want {
				t.Fatalf("Calculate() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func groupPod(group, worker string, phase corev1.PodPhase, ready bool, waitingReason string) corev1.Pod {
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
		"leaderworkerset.sigs.k8s.io/group-index":  group,
		"leaderworkerset.sigs.k8s.io/worker-index": worker,
	}}, Status: corev1.PodStatus{Phase: phase}}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	if waitingReason != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: waitingReason},
		}}}
	}
	return pod
}
