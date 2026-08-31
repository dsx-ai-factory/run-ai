/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v2alpha1

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
)

// run:ai rules are static so calculating them once per service is sufficient
var runaiRules *InferenceSpec

// run:ai rules are general guidelines that we apply on top of the admin policy
// we use them in two kinds of validations
//   - the (admin) policy is validated against the run:ai rules
//   - the (researcher) workload is validated against both (admin) policy and runai:rules
//
// NOTE!
// default values here are mainly informatory or used by the UI. that is because the
// run:ai rules are not "merged" to the merge stack, but work aside. it is the duty
// of the builders to make sure the default are applied. this is not perfect, but since
// the run:ai rules are mainly for enforcement of limits/formats and less for default,
// I did not bother to go all the way with values.
func GetRunaiRules(ctx context.Context, nodePoolLister field_groups.NodePoolNameLister) *InferenceSpec {
	workloadRunaiRules := field_groups.WorkloadRunaiRules()
	if runaiRules == nil {
		runaiRules = &InferenceSpec{
			WorkloadFields:           inferenceWorkloadRules(*workloadRunaiRules.DeepCopy()),
			ResourceAllocationFields: field_groups.ResourceAllocationRunaiRules,
			InferenceFields:          field_groups.InferenceRunaiRules,
			SchedulingFields:         field_groups.RunaiSchedulingRules(),
		}
	}

	// either way, update the scheduler runai rules against the current situation of the cluster
	runaiRules.RunaiSchedulerFields = field_groups.RunaiSchedulerRunaiRules(ctx, nodePoolLister)

	return runaiRules
}
