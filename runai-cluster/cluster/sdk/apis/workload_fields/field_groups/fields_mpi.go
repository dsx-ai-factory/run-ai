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
type MpiFields struct {

	//   This workload produces mpijob
	Mpi *flds.BoolField `json:"mpi,omitempty"`

	//   Number of distributed training processes that will be allocated for the created mpijob.
	MpiProcs *flds.IntField `json:"processes,omitempty"`

	//   Number of slots to allocate per worker in the created mpijob.
	SlotsPerWorker *flds.IntField `json:"slotsPerWorker,omitempty"`
}

var MpiRunaiRules = MpiFields{
	MpiProcs: &flds.IntField{
		Value: ptr.To(constants.DefaultWorkers), // RUNAI DEFAULT
		Rules: &flds.IntRules{
			Min: ptr.To(constants.MinWorkers),
		},
	},
}
