/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package unschedulable

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	v1 "k8s.io/api/core/v1"
)

func ParseExceedingResources(reason v2alpha2.UnschedulableExplanation) (v1.ResourceList, error) {
	var (
		requested v1.ResourceList
		used      v1.ResourceList
		capacity  v1.ResourceList
	)
	switch reason.Reason {
	case v2alpha2.NonPreemptibleOverQuota:
		requested = reason.Details.QueueDetails.PodGroupRequestedNonPreemptibleResources
		used = reason.Details.QueueDetails.QueueAllocatedNonPreemptibleResources
		capacity = reason.Details.QueueDetails.QueueDeservedResources
	case v2alpha2.OverLimit:
		requested = reason.Details.QueueDetails.PodGroupRequestedResources
		used = reason.Details.QueueDetails.QueueAllocatedResources
		capacity = reason.Details.QueueDetails.QueueResourceLimits
	default:
		return nil, fmt.Errorf("unsupported reason %s", reason.Reason)
	}

	totalRequest := resources.SumResources(requested, used)

	return resources.GetExceedingResources(totalRequest, capacity), nil
}
