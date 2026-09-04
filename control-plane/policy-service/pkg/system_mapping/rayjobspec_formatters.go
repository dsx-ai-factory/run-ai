/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemRayJobSpecFormatterFields returns custom formatter fields for
// ray.io/RayJob/v1 workload-level fields that need enum validation.
func BuildSystemRayJobSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	addRayJobSpecField(m, "submissionMode", `{
	  "submissionMode": $ARGS.named.value
	}`, nil,
		map[string][]string{
			"value": {"K8sJobMode", "HTTPMode", "InteractiveMode", "SidecarMode"},
		})

	return m
}

func addRayJobSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr: &setExpr,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
