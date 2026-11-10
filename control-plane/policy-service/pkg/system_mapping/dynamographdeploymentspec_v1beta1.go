/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const (
	SystemDynamoGraphDeploymentV1beta1SpecFgName        = "dynamoGraphDeploymentV1beta1Spec"
	SystemDynamoGraphDeploymentV1beta1SpecFormatterName = "runai-dynamoGraphDeploymentV1beta1Spec"
)

func BuildSystemDynamoGraphDeploymentV1beta1SpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields, selectors := BuildSystemDynamoGraphDeploymentSpecFg()
	backend := fields["backendFramework"]
	backend.Formatter = SystemDynamoGraphDeploymentV1beta1SpecFormatterName
	fields["backendFramework"] = backend
	return fields, selectors
}

func BuildSystemDynamoGraphDeploymentV1beta1SpecFormatterFields() map[string]api.FormatterField {
	return map[string]api.FormatterField{
		"backendFramework": {
			GetExpr:  convutils.Ref(`.backendFramework`),
			SetExpr:  convutils.Ref(`{"backendFramework": $ARGS.named.value}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"value": {"vllm", "sglang", "trtllm"}},
		},
	}
}
