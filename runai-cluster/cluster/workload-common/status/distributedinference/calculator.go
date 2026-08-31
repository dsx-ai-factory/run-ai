/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package distributedinference calculates distributed inference workload status without Kubernetes I/O.
package distributedinference

import (
	"fmt"
	"strings"

	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	corev1 "k8s.io/api/core/v1"
)

const (
	groupIndexLabel  = "leaderworkerset.sigs.k8s.io/group-index"
	workerIndexLabel = "leaderworkerset.sigs.k8s.io/worker-index"
)

type Input struct {
	Exists            bool
	Replicas          int32
	GroupSize         int
	Pods              []corev1.Pod
	BuildError        *string
	MissingResources  []string
	MNNVLEnabled      bool
	MNNVLReady        bool
	MarkedForDeletion bool
}

type RuntimeInput struct {
	Replicas  int32
	GroupSize int
	Pods      []corev1.Pod
}

type WorkloadState struct {
	Exists            bool
	BuildError        *string
	MissingResources  []string
	MNNVLEnabled      bool
	MNNVLReady        bool
	MarkedForDeletion bool
}

type Result struct {
	Phase         statuscontract.Phase
	Message       string
	Replicas      int32
	ReadyReplicas int32
}

type podPhase string

const (
	podPhasePending      podPhase = "Pending"
	podPhaseRunning      podPhase = "Running"
	podPhaseReady        podPhase = "Ready"
	podPhaseFailed       podPhase = "Failed"
	podPhaseInitializing podPhase = "Initializing"
)

type phaseInfo struct {
	replicas     int32
	pending      int32
	initializing int32
	ready        int32
	failed       int32
}

func Calculate(input Input) Result {
	base := Result{}
	if input.Exists {
		base, _ = CalculateRuntime(RuntimeInput{
			Replicas: input.Replicas, GroupSize: input.GroupSize, Pods: input.Pods,
		})
	}
	return ApplyWorkloadState(base, WorkloadState{
		Exists:            input.Exists,
		BuildError:        input.BuildError,
		MissingResources:  input.MissingResources,
		MNNVLEnabled:      input.MNNVLEnabled,
		MNNVLReady:        input.MNNVLReady,
		MarkedForDeletion: input.MarkedForDeletion,
	})
}

func CalculateRuntime(input RuntimeInput) (Result, error) {
	info := phaseInfo{replicas: input.Replicas}
	groups := groupPods(input.Pods)
	var calculationErr error
	groupSize := input.GroupSize
	if groupSize <= 0 {
		groupSize = 1
	}

	for _, group := range groups {
		phase, err := determineGroupPhase(group, groupSize)
		if err != nil && calculationErr == nil {
			calculationErr = err
		}
		switch phase {
		case statuscontract.Pending:
			info.pending++
		case statuscontract.Initializing:
			info.initializing++
		case statuscontract.Ready:
			info.ready++
		case statuscontract.Failed:
			info.failed++
		}
	}
	if missing := input.Replicas - int32(len(groups)); missing > 0 {
		info.pending += missing
	}

	return Result{
		Phase:         aggregatePhase(info),
		Message:       phaseMessage(info),
		Replicas:      input.Replicas,
		ReadyReplicas: info.ready,
	}, calculationErr
}

func ApplyWorkloadState(base Result, state WorkloadState) Result {
	if state.MarkedForDeletion {
		base.Phase = statuscontract.Deleting
		base.Message = ""
		return base
	}
	if !state.Exists {
		base.Phase = statuscontract.Creating
		base.Message = ""
		if state.BuildError != nil && strings.Contains(*state.BuildError, "LeaderWorkerSet") {
			base.Phase = statuscontract.Failed
		}
		return base
	}

	hasIssue := false
	if state.BuildError != nil {
		base.Message = appendMessage(base.Message, *state.BuildError)
		hasIssue = true
	} else if len(state.MissingResources) > 0 {
		base.Message = appendMessage(base.Message, "Missing resources: "+strings.Join(state.MissingResources, ", "))
		hasIssue = true
	}
	if state.MNNVLEnabled && !state.MNNVLReady {
		base.Message = appendMessage(base.Message, mnnvltypes.UserFacingFeatureName+" is not ready")
		hasIssue = true
	}
	if base.Phase == statuscontract.Ready && hasIssue {
		base.Phase = statuscontract.Degraded
	}
	return base
}

func groupPods(pods []corev1.Pod) map[string]map[string]podPhase {
	groups := make(map[string]map[string]podPhase)
	for i := range pods {
		group, groupOK := pods[i].Labels[groupIndexLabel]
		worker, workerOK := pods[i].Labels[workerIndexLabel]
		if !groupOK || !workerOK {
			continue
		}
		if groups[group] == nil {
			groups[group] = make(map[string]podPhase)
		}
		groups[group][worker] = determinePodPhase(&pods[i])
	}
	return groups
}

func determineGroupPhase(group map[string]podPhase, expectedSize int) (statuscontract.Phase, error) {
	if len(group) < expectedSize {
		return statuscontract.Pending, nil
	}
	for _, phase := range group {
		if phase == podPhaseFailed {
			return statuscontract.Failed, nil
		}
	}
	for _, phase := range group {
		if phase == podPhasePending {
			return statuscontract.Pending, nil
		}
	}
	for _, phase := range group {
		if phase == podPhaseInitializing {
			return statuscontract.Initializing, nil
		}
	}
	leader, ok := group["0"]
	if !ok {
		return statuscontract.Failed, fmt.Errorf("leader pod not found")
	}
	if leader == podPhaseReady {
		return statuscontract.Ready, nil
	}
	return statuscontract.Initializing, nil
}

func determinePodPhase(pod *corev1.Pod) podPhase {
	if pod.DeletionTimestamp != nil {
		return podPhaseFailed
	}
	switch pod.Status.Phase {
	case corev1.PodRunning:
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return podPhaseReady
			}
		}
		return podPhaseRunning
	case corev1.PodSucceeded:
		return podPhaseRunning
	case corev1.PodFailed:
		return podPhaseFailed
	case corev1.PodPending:
		statuses := append([]corev1.ContainerStatus{}, pod.Status.ContainerStatuses...)
		statuses = append(statuses, pod.Status.InitContainerStatuses...)
		for _, status := range statuses {
			if status.State.Waiting == nil {
				continue
			}
			switch status.State.Waiting.Reason {
			case "ContainerCreating", "PodInitializing", "ImagePullBackOff", "ErrImagePull", "ErrImageNeverPull":
				return podPhaseInitializing
			case "CreateContainerError", "CreateContainerConfigError":
				return podPhaseFailed
			}
		}
		return podPhasePending
	default:
		return podPhasePending
	}
}

func aggregatePhase(info phaseInfo) statuscontract.Phase {
	if info.replicas == 0 {
		return statuscontract.Ready
	}
	if info.failed > 0 {
		if info.ready > 0 {
			return statuscontract.Degraded
		}
		return statuscontract.Failed
	}
	if info.pending == info.replicas {
		return statuscontract.Pending
	}
	if info.ready != info.replicas {
		return statuscontract.Initializing
	}
	return statuscontract.Ready
}

func phaseMessage(info phaseInfo) string {
	message := fmt.Sprintf("Replicas: %d", info.replicas)
	for _, item := range []struct {
		name  string
		count int32
	}{{"Ready", info.ready}, {"Initializing", info.initializing}, {"Pending", info.pending}, {"Failed", info.failed}} {
		if item.count > 0 {
			message += fmt.Sprintf("; %s: %d", item.name, item.count)
		}
	}
	return message
}

func appendMessage(base, suffix string) string {
	if suffix == "" {
		return base
	}
	if base == "" {
		return suffix
	}
	return base + "; " + suffix
}
