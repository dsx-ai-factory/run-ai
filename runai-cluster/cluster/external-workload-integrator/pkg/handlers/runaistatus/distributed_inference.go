/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	"sort"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	commondistributedinference "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/distributedinference"
)

type DistributedInferenceInput = commondistributedinference.Input

type DistributedInferenceResult struct {
	Phase         v1alpha1.ExternalWorkloadPhase
	Message       string
	Replicas      int32
	ReadyReplicas int32
}

func ComputeDistributedInferenceStatus(input DistributedInferenceInput,
	networkingPhase dromosv1alpha1.PluginPhase,
) DistributedInferenceResult {
	input.MissingResources = append([]string(nil), input.MissingResources...)
	sort.Strings(input.MissingResources)
	result := commondistributedinference.Calculate(input)

	phase := v1alpha1.ExternalWorkloadPhase(result.Phase)
	if result.Phase == statuscontract.Ready {
		phase = v1alpha1.ExWRunningPhase
	}
	output := DistributedInferenceResult{
		Phase:         phase,
		Message:       result.Message,
		Replicas:      result.Replicas,
		ReadyReplicas: result.ReadyReplicas,
	}

	// The phase groups mirror WorkloadRunner plugin aggregation, scoped deliberately to
	// gating an otherwise Running EWI lifecycle result on networking readiness.
	if output.Phase != v1alpha1.ExWRunningPhase {
		return output
	}
	switch networkingPhase {
	case dromosv1alpha1.PluginPhaseFailed:
		output.Phase = v1alpha1.ExWDegradedPhase
	case dromosv1alpha1.PluginPhaseCreating,
		dromosv1alpha1.PluginPhasePending,
		dromosv1alpha1.PluginPhaseInitializing,
		dromosv1alpha1.PluginPhaseUnknown:
		output.Phase = v1alpha1.ExWInitializingPhase
	}
	return output
}
