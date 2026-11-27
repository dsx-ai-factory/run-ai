/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"

const dynamoParallelRestartStrategy = "Parallel"

func BuildSystemDynamoGraphDeploymentV1beta1SpecFormatterFields() map[string]api.FormatterField {
	fields := make(map[string]api.FormatterField)
	addDynamoFormatterField(fields, dynamoBackendFrameworkField, `{
  "backendFramework": $ARGS.named.value
 }`, nil, map[string][]string{
		dynamoFormatterValueArg: {"vllm", "sglang", "trtllm"},
	})
	addDynamoFormatterField(fields, dynamoRestartStrategyTypeField, `{
  "restart": {"strategy": {"type": $ARGS.named.value}}
 }`, nil, map[string][]string{
		dynamoFormatterValueArg: {"Sequential", dynamoParallelRestartStrategy},
	})
	return fields
}
