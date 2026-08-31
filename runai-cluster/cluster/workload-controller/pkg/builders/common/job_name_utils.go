/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

func GetJobType(workload *jobs.JobWorkload) string {
	switch workload.Kind {
	case constants.KindDistributedWorkload:
		return workload.Spec.JobType
	case constants.KindInferenceWorkload:
		return constants.KnativeServiceKind
	case constants.KindInteractiveWorkload:
		return constants.KindRunaiJob
	case constants.KindTrainingWorkload:
		return constants.KindRunaiJob

	default:
		return workload.Kind
	}
}

func CalculatePriorityClassNameLabel(specPriorityClassName string, preemptibility *flds.StringField, workloadKind string) string {
	if specPriorityClassName != "" {
		return specPriorityClassName
	}

	if isPreemptibleSpecRelevant(workloadKind) {
		if IsPreemptiblitySet(preemptibility) {
			return constants.InteractivePreemptiblePriorityClass
		}
		return constants.InteractivePriorityClass
	}

	return ""
}

func IsPreemptiblitySet(preemptibility *flds.StringField) bool {
	return preemptibility.Deref() != "" && *preemptibility.Value == field_groups.Preemptible
}

func isPreemptibleSpecRelevant(workloadKind string) bool {
	return workloadKind == constants.KindInteractiveWorkload
}

func CalculatePreemptibilityNameLabel(preemptibility *flds.StringField, workloadKind string) string {
	if preemptibility != nil && preemptibility.Value != nil {
		return *preemptibility.Value
	}

	// For backward compatibility with the deprecated preemptible field:
	// Only interactive workloads default were explicitly set to NonPreemptible
	if workloadKind == constants.KindInteractiveWorkload {
		return string(field_groups.NonPreemptible)
	}

	return ""
}
