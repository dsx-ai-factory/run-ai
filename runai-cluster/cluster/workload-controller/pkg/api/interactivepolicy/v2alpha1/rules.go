/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v2alpha1

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
)

// run:ai rules are constant, thus calculating them once per service is sufficient
var runaiRules *InteractiveSpec

// run:ai rules are general guidelines that we apply on top of the admin policy
// we use them in two kinds of validations
//   - each (admin) policy is validated against the run:ai rules
//   - each (researcher) workload is validated against both admin policy and run:ai rules
//
// NOTE!
// default values here are mainly informatory or used by the UI. that is because the
// run:ai rules are not "merged" to the merge stack, but work aside. it is the duty
// of the builders to make sure the default are applies. this is not perfect, but since
// the run:ai rules are mainly for enforcement of limits/formats and less for default,
// I did not bother to go all the way with values.
func GetRunaiRules(ctx context.Context, nodePoolLister field_groups.NodePoolNameLister) *InteractiveSpec {
	if runaiRules == nil {

		runaiRules = &InteractiveSpec{
			WorkloadFields:           field_groups.WorkloadRunaiRules(),
			ResourceAllocationFields: field_groups.ResourceAllocationRunaiRules,
			MpiFields:                field_groups.MpiRunaiRules,
			SchedulingFields:         field_groups.RunaiSchedulingRules(),
		}
	}

	// either way, update the scheduler runai rules against the current situation of the cluster
	runaiRules.RunaiSchedulerFields = field_groups.RunaiSchedulerRunaiRules(ctx, nodePoolLister)

	return runaiRules
}
