/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
)

// GetDistributedInferenceWorkerSpec returns the worker spec from the workload, falling back to leader spec if no worker spec is defined.
// In LWS, WorkerTemplate is mandatory, even if there are no workers (namely, Size = 1, only leader).
// A worker StatefulSet will always be created, even if there are no workers (namely, Size = 1, only leader). In this case, the StatefulSet will
// exist but will be of size 0.
func GetDistributedInferenceWorkerSpec(workload *dinfw.DistributedInferenceWorkload) *dinfp.DistributedInferenceEntitySpec {
	if workload.WorkerSpec != nil {
		return workload.WorkerSpec
	}
	return &workload.LeaderSpec
}

// GetUniqueLwsName returns a unique name for the LWS, based on the workload name and UID.
func GetUniqueLwsName(workload *dinfw.DistributedInferenceWorkload) string {
	// lws adds "-0", "-1-1", etc. to the pod names, which is problematic as it can lead to conflicts.
	// for example, workload "a" and workload "a-1" will both have a pod named "a-1-1", but for the first workload
	// it will be the worker, while for the second workload it will be the leader.
	// we add a unique suffix to avoid that. we take the last characters of the workload ID to ensure uniqueness.
	// having a unique LWS name results in unique pod names.

	LwsNameSuffixLength := 4

	workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
	if len(workloadId) < LwsNameSuffixLength {
		// fallback for testing, where workload ID is empty
		return workload.Name
	}

	return workload.Name + "-" + workloadId[len(workloadId)-LwsNameSuffixLength:]
}
