/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemRayJobSpecFg builds the field group for ray.io/RayJob/v1
// workload-level scalars under .spec (excluding the embedded rayClusterSpec,
// which uses the shared SystemRayClusterSpecFgName).
func BuildSystemRayJobSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(rayJobSpecFgFields())
	return fields, nil
}

func rayJobSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	keyValueMap := dbmodel.BuiltinFormatterKeyValueMap
	rayJobSpec := SystemRayJobSpecFormatterName

	return []SystemFgField{
		{
			Name:      "entrypoint",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "runtimeEnvYAML",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "shutdownAfterJobFinishes",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "ttlSecondsAfterFinished",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "activeDeadlineSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "backoffLimit",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "entrypointNumCpus",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_NUMBER,
		},
		{
			Name:      "entrypointNumGpus",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_NUMBER,
		},
		{
			Name:      "entrypointResources",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "submissionMode",
			Formatter: rayJobSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "deletionPolicy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "managedBy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "metadata",
			Formatter: keyValueMap,
			SpecRef:   strRef("metadata"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "clusterSelector",
			Formatter: keyValueMap,
			SpecRef:   strRef("clusterSelector"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "submitterConfig-BackoffLimit",
			Formatter: keyValue,
			SpecRef:   strRef("submitterConfig.backoffLimit"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "suspend",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
	}
}
