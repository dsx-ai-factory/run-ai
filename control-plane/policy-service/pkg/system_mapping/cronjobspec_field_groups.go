/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemCronJobSpecFg builds the system-scoped "cronJobSpec" field group,
// covering the CronJob-specific fields at the CronJob .spec level (above the
// job template). Field paths are rooted at the workload .spec.
//
// concurrencyPolicy's enum ("Allow" | "Forbid" | "Replace") is enforced at
// write time via the runai-cronJobSpec custom formatter.
func BuildSystemCronJobSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(cronJobSpecFgFields())
	return fields, map[string]api.ItemSelector{}
}

func cronJobSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	cronJobSpec := SystemCronJobSpecFormatterName

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// Plain scalars
		////////////////////////////////////////////////////////
		{
			Name:      "schedule",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "timeZone",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "suspend",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "startingDeadlineSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "successfulJobsHistoryLimit",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "failedJobsHistoryLimit",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},

		////////////////////////////////////////////////////////
		// Enum scalar (custom formatter for enum enforcement)
		////////////////////////////////////////////////////////
		{
			Name:      "concurrencyPolicy",
			Formatter: cronJobSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
