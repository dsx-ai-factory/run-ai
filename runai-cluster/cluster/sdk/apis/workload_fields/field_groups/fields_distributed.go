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

// DistributedFields specifies the values and rules for distributed related fields.
// Following distributed jobs are supported:
//   - MPIJob
//   - PyTorchJob
//
// It is validated that details for exactly one job are provided.
// +k8s:deepcopy-gen=true
type DistributedFields struct {
	// The type of distributed job.
	//+kubebuilder:validation:Enum:=MPIJob;PyTorchJob;TFJob;XGBoostJob;JAXJob
	JobType string `json:"jobType,omitempty"`

	// The desired number of worker pods.
	Workers *flds.IntField `json:"workers,omitempty"`

	// Request the job not to have a master pod.
	NoMaster *flds.BoolField `json:"noMaster,omitempty"`

	// RunPolicy is shared between all distributed jobs.
	RunPolicy *flds.RunPolicyField `json:"runPolicy,omitempty"`

	// Request the job to be non-preemptible.
	NonPreemptible *flds.BoolField `json:"nonPreemptible,omitempty"`

	// Specific fields for distributed MPI Job.
	MpiJob *flds.MpiField `json:"mpiJob,omitempty"`

	// Specific fields for distributed PyTorch Job.
	PyTorchJob *flds.PyTorchField `json:"pyTorchJob,omitempty"`

	// Specify by which controller the workload is managed by, default empty string means workload controller
	// +kubebuilder:default=""
	ManagedBy string `json:"managedBy,omitempty"`
}

var DistributedRunaiRules = DistributedFields{
	Workers: &flds.IntField{
		Value: ptr.To(constants.DefaultWorkers), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(constants.MinWorkers),
		},
	},

	RunPolicy: &flds.RunPolicyField{
		Values: &flds.RunPolicyMembers{
			CleanPodPolicy: &flds.StringField{
				Rules: &flds.StringRules{
					Options: &[]flds.OptionRule{
						{
							Name:  ptr.To("None"),
							Value: ptr.To("None"),
						},
						{
							Name:  ptr.To("All"),
							Value: ptr.To("All"),
						},
						{
							Name:  ptr.To("Running"),
							Value: ptr.To("Running"),
						},
					},
				},
			},
		},
	},

	MpiJob: &flds.MpiField{
		Values: &flds.MpiMembers{
			SlotsPerWorker: &flds.IntField{
				Value: ptr.To(constants.DefaultSlotsPerWorker), // RUNAI DEFAULT
				Rules: &flds.IntRules{
					Min: ptr.To(constants.MinSlotsPerWorker),
				},
			},
			SshAuthMountPath: &flds.StringField{
				Rules: &flds.StringRules{
					Required: ptr.To(false),
					CanEdit:  ptr.To(true),
				}},
		},
	},

	PyTorchJob: &flds.PyTorchField{
		Values: &flds.PyTorchMembers{
			MinReplicas: &flds.IntField{
				Rules: &flds.IntRules{
					Min: ptr.To(constants.MinWorkers),
				},
			},
			MaxReplicas: &flds.IntField{
				Rules: &flds.IntRules{
					Min: ptr.To(constants.MinWorkers),
				},
			},
		},
	},
}
