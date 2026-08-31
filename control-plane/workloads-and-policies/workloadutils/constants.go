/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadutils

import (
	workloadApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type NativeWorkloadConfig struct {
	Category       string
	Priority       string
	Preemptibility workloadApi.WorkloadTypePreemptibility
	WorkloadKind   string
	WorkloadType   string
	Version        string
}

var DefaultRunaiWorkloadTypesConfig = map[string]NativeWorkloadConfig{
	"Training": {
		Category:       "Train",
		Priority:       "low",
		Preemptibility: workloadApi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		WorkloadKind:   "TrainingWorkload",
		WorkloadType:   "Training",
		Version:        "v2alpha1",
	},
	"Distributed": {
		Category:       "Train",
		Priority:       "low",
		Preemptibility: workloadApi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
		WorkloadKind:   "DistributedWorkload",
		WorkloadType:   "Distributed",
		Version:        "v2alpha1",
	},
	"Workspace": {
		Category:       "Build",
		Priority:       "high",
		Preemptibility: workloadApi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		WorkloadKind:   "InteractiveWorkload",
		WorkloadType:   "Workspace",
		Version:        "v2alpha1",
	},
	"Inference": {
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: workloadApi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		WorkloadKind:   "InferenceWorkload",
		WorkloadType:   "Inference",
		Version:        "v2alpha1",
	},
	"DistributedInference": {
		Category:       "Deploy",
		Priority:       "very-high",
		Preemptibility: workloadApi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE,
		WorkloadKind:   "DistributedInferenceWorkload",
		WorkloadType:   "DistributedInference",
		Version:        "v1alpha1",
	},
}
