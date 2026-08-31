/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemRayClusterSpecFormatterFields returns custom formatter fields for
// RayClusterSpec enum-validated fields. Bound to the shared rayClusterSpec
// field group; reused by RayCluster.spec, RayJob.spec.rayClusterSpec, and
// RayService.spec.rayClusterConfig.
func BuildSystemRayClusterSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	addRayClusterSpecField(m, "autoscalerOptions-UpscalingMode", `{
	  "autoscalerOptions": {
		"upscalingMode": $ARGS.named.value
	  }
	}`, nil,
		map[string][]string{
			"value": {"Default", "Aggressive", "Conservative"},
		})

	addRayClusterSpecField(m, "autoscalerOptions-ImagePullPolicy", `{
	  "autoscalerOptions": {
		"imagePullPolicy": $ARGS.named.value
	  }
	}`, nil,
		map[string][]string{
			"value": {"Always", "IfNotPresent", "Never"},
		})

	addRayClusterSpecField(m, "autoscalerOptions-Version", `{
	  "autoscalerOptions": {
		"version": $ARGS.named.value
	  }
	}`, nil,
		map[string][]string{
			"value": {"v1", "v2"},
		})

	return m
}

func addRayClusterSpecField(
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
