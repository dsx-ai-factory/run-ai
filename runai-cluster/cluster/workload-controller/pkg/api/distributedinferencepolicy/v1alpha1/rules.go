/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
)

// run:ai rules are static so calculating them once per service is sufficient
var runaiRules *DistributedInferenceSpec
var runaiRulesForEntity *DistributedInferenceEntitySpec

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
func GetRunaiRules(ctx context.Context, nodePoolLister field_groups.NodePoolNameLister) *DistributedInferenceSpec {
	workloadRunaiRules := field_groups.WorkloadRunaiRules()
	if runaiRules == nil {
		runaiRules = &DistributedInferenceSpec{
			IdentityFields:                      workloadRunaiRules.IdentityFields,
			DistributedInferenceFields:          field_groups.DistributedInferneceRunaiRules,
			DistributedInferenceSchedulerFields: field_groups.DistributedInferenceSchedulerRunaiRules(ctx, nodePoolLister),
			SchedulingFields:                    field_groups.RunaiSchedulingRules(),
		}
	}

	return runaiRules
}

func GetRunaiRulesForEntity() *DistributedInferenceEntitySpec {
	if runaiRulesForEntity == nil {
		runaiRulesForEntity = &DistributedInferenceEntitySpec{
			ContainerFields:          field_groups.RunaiContainerRules(),
			StorageFields:            field_groups.StorageRunaiRules,
			ResourceAllocationFields: field_groups.ResourceAllocationRunaiRules,
		}
	}

	// either way, update the scheduler runai rules against the current situation of the cluster
	runaiRulesForEntity.DistributedInferenceEntitySchedulerFields = field_groups.DistributedInferenceEntitySchedulerRunaiRules()

	return runaiRulesForEntity
}
