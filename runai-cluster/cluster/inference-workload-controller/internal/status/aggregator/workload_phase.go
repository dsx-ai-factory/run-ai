/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"

	commoninference "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/inference"
	corev1 "k8s.io/api/core/v1"
)

func (sc *StatusCalculator) handlePhaseAndPhaseMessage(_ context.Context) {
	input := commoninference.Input{
		Service:           sc.inferenceServiceState(),
		Pods:              sc.latestCreatedRevisionPods,
		BuildError:        sc.buildError,
		MarkedForDeletion: sc.workload.IsMarkedForDeletion(),
	}
	result := commoninference.Calculate(input)

	log.V(1).Info("new workload phase",
		"workloadName", sc.workload.Name,
		"newPhase", result.Phase,
		"namespace", sc.workload.Namespace)

	sc.workload.Status.InferenceStatus.Phase = string(result.Phase)
	sc.workload.Status.InferenceStatus.PhaseMessage = result.Message
}

func (sc *StatusCalculator) inferenceServiceState() *commoninference.ServiceState {
	state := &commoninference.ServiceState{}
	if sc.knativeService == nil {
		state.Conditions = []commoninference.ServiceCondition{{Status: corev1.ConditionFalse}}
	} else {
		state.LatestCreatedRevision = sc.knativeService.Status.LatestCreatedRevisionName
		state.LatestReadyRevision = sc.knativeService.Status.LatestReadyRevisionName
		for _, condition := range sc.knativeService.Status.Conditions {
			state.Conditions = append(state.Conditions, commoninference.ServiceCondition{
				Status:  condition.Status,
				Message: condition.Message,
			})
		}
	}

	if sc.workload.Status.LatestCreatedRevisionName != nil {
		state.LatestCreatedRevision = *sc.workload.Status.LatestCreatedRevisionName
	}
	if sc.workload.Status.LatestReadyRevisionName != nil {
		state.LatestReadyRevision = *sc.workload.Status.LatestReadyRevisionName
	}
	return state
}
