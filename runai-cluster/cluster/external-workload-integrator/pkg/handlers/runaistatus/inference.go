/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaistatus

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/contract"
	commoninference "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/status/inference"
)

type InferenceServiceCondition = commoninference.ServiceCondition
type InferenceServiceState = commoninference.ServiceState
type InferenceInput = commoninference.Input

// ComputeInferenceStatus shows a Running workload as Degraded while
// snapshot-plugin reports Failed, as distributed inference does for a failed
// networking plugin: it serves, but its replicas start cold. The plugin's issue
// says why. snapshot is nil for a runner without the plugin's status.
func ComputeInferenceStatus(input InferenceInput,
	snapshot *dromosv1alpha1.PluginStatus,
) (v1alpha1.ExternalWorkloadPhase, string) {
	result := commoninference.Calculate(input)
	if result.Phase != statuscontract.Ready {
		return v1alpha1.ExternalWorkloadPhase(result.Phase), result.Message
	}
	if snapshot == nil || snapshot.Phase != dromosv1alpha1.PluginPhaseFailed {
		return v1alpha1.ExWRunningPhase, result.Message
	}
	if len(snapshot.Issues) > 0 {
		return v1alpha1.ExWDegradedPhase, snapshot.Issues[0].Description
	}
	return v1alpha1.ExWDegradedPhase, result.Message
}
