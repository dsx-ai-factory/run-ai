/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type LifeCycleFields struct {

	// Used with Hyperparameter Optimization. Specifies the number of successful pods the job should reach to be completed.
	// The Job will be marked as successful once the specified amount of pods has succeeded. The default value for 'completions' is 1.
	// The 'parallelism' flag should be smaller or equal to 'completions'
	// For more information see: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	Completions *flds.IntField `json:"completions,omitempty"`

	// Specifies the maximum desired number of pods the workload should
	// run at any given time. The actual number of pods running in a steady state will
	// be less than this number when ((.spec.completions - .status.successful) < .spec.parallelism),
	// i.e. when the work left to do is less than max parallelism.
	// For more information, see: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	Parallelism *flds.IntField `json:"parallelism,omitempty"`
}

var LifeCycleRunaiRules = LifeCycleFields{
	Completions: &flds.IntField{
		Value: ptr.To(constants.DefaultCompletions), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(0),
		},
	},

	Parallelism: &flds.IntField{
		Value: ptr.To(constants.DefaultParallelism), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(0),
		},
	},
}
