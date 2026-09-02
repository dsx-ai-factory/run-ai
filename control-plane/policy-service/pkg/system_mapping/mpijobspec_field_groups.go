/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemMPIJobSpecFg builds the field group for kubeflow.org/MPIJob/v1
// workload-level fields. This includes:
//   - slotsPerWorker (MPIJob-specific)
//
// Note: runPolicy fields are provided by the shared kubeflowRunPolicy field group.
func BuildSystemMPIJobSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(mpiJobSpecFgFields())
	return fields, nil
}

func mpiJobSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// MPI-specific fields
		////////////////////////////////////////////////////////
		{
			Name:      "slotsPerWorker",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
	}
}
