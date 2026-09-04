/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type DistributedInferenceFields struct {
	// The number of workers to run (per replica).
	Workers *flds.IntField `json:"workers,omitempty"`

	// The number of replicas to run.
	Replicas *flds.IntField `json:"replicas,omitempty"`

	// The serving ports to expose, for model access.
	ServingPorts *flds.ServingPortsField `json:"servingPorts,omitempty"`

	// The startup policy for the worker pods. Options are: LeaderCreated (default), LeaderReady.
	StartupPolicy *flds.StringField `json:"startupPolicy,omitempty"`

	// The restart policy when pod failures happen. Options are: RecreateGroupOnPodRestart (default), None.
	RestartPolicy *flds.StringField `json:"restartPolicy,omitempty"`
}

var DistributedInferneceRunaiRules = DistributedInferenceFields{
	Workers: &flds.IntField{
		Value: ptr.To(constants.DefaultDistributedInferenceWorkers), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(constants.MinDistributedInferenceWorkers),
		},
	},
	Replicas: &flds.IntField{
		Value: ptr.To(constants.DefaultDistributedInferenceReplicas), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(constants.MinDistributedInferenceReplicas),
		},
	},
	StartupPolicy: &flds.StringField{
		Value: ptr.To(constants.DefaultDistributedInferenceStartupPolicy), // RUNAI DEFAULT
		Rules: &flds.StringRules{
			Options: &[]flds.OptionRule{
				{
					Value: ptr.To(constants.DistributedInferenceStartupPolicyLeaderCreated),
					Name:  ptr.To(constants.DistributedInferenceStartupPolicyLeaderCreated),
				},
				{
					Value: ptr.To(constants.DistributedInferenceStartupPolicyLeaderReady),
					Name:  ptr.To(constants.DistributedInferenceStartupPolicyLeaderReady),
				},
			},
		},
	},
	RestartPolicy: &flds.StringField{
		Value: ptr.To(constants.DefaultDistributedInferenceRestartPolicy), // RUNAI DEFAULT
		Rules: &flds.StringRules{
			Options: &[]flds.OptionRule{
				{
					Value: ptr.To(constants.DistributedInferenceRestartPolicyRecreateGroupOnPodRestart),
					Name:  ptr.To(constants.DistributedInferenceRestartPolicyRecreateGroupOnPodRestart),
				},
				{
					Value: ptr.To(constants.DistributedInferenceRestartPolicyNone),
					Name:  ptr.To(constants.DistributedInferenceRestartPolicyNone),
				},
			},
		},
	},
}
