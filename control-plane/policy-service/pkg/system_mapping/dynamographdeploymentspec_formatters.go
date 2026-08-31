/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemDynamoGraphDeploymentSpecFormatterFields returns the formatter fields for DynamoGraphDeployment.
func BuildSystemDynamoGraphDeploymentSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	// Workload-level fields
	addDynamoFormatterField(m, "backendFramework", `{
	  "backendFramework": $ARGS.named.value
	}`, nil, map[string][]string{
		"value": {"vllm", "sglang", "trtllm", "tensorrt-llm", "nixl", "dynamo"},
	})

	addDynamoFormatterField(m, "restart-id", `{
	  "restart": {"id": $ARGS.named.value}
	}`, nil, nil)

	addDynamoFormatterField(m, "restart-strategy-type", `{
	  "restart": {"strategy": {"type": $ARGS.named.value}}
	}`, nil, map[string][]string{
		"value": {"Sequential", "Parallel"},
	})

	// Service-level fields
	addDynamoFormatterField(m, "replicas", `{
	  "replicas": ($ARGS.named.value | tonumber)
	}`, nil, nil)

	addDynamoFormatterField(m, "componentType", `{
	  "componentType": $ARGS.named.value
	}`, nil, nil)

	addDynamoFormatterField(m, "subComponentType", `{
	  "subComponentType": $ARGS.named.value
	}`, nil, nil)

	addDynamoFormatterField(m, "dynamoNamespace", `{
	  "dynamoNamespace": $ARGS.named.value
	}`, nil, nil)

	addDynamoFormatterField(m, "globalDynamoNamespace", `{
	  "globalDynamoNamespace": (if ($ARGS.named.value | type) == "boolean" then $ARGS.named.value else ($ARGS.named.value | test("^(true|1)$"; "i")) end)
	}`, nil, nil)

	addDynamoFormatterField(m, "serviceName", `{
	  "serviceName": $ARGS.named.value
	}`, nil, nil)

	addDynamoFormatterField(m, "sharedMemory-size", `{
	  "sharedMemory": {"size": $ARGS.named.value}
	}`, nil, nil)

	addDynamoFormatterField(m, "sharedMemory-disabled", `{
	  "sharedMemory": {"disabled": (if ($ARGS.named.value | type) == "boolean" then $ARGS.named.value else ($ARGS.named.value | test("^(true|1)$"; "i")) end)}
	}`, nil, nil)

	// Checkpoint fields
	addDynamoFormatterField(m, "checkpoint-enabled", `{
	  "checkpoint": {"enabled": (if ($ARGS.named.value | type) == "boolean" then $ARGS.named.value else ($ARGS.named.value | test("^(true|1)$"; "i")) end)}
	}`, nil, nil)

	addDynamoFormatterField(m, "checkpoint-checkpointRef", `{
	  "checkpoint": {"checkpointRef": $ARGS.named.value}
	}`, nil, nil)

	// Multinode fields
	addDynamoFormatterField(m, "multinode-nodeCount", `{
	  "multinode": {"nodeCount": ($ARGS.named.value | tonumber)}
	}`, nil, nil)

	// Ingress fields
	addDynamoFormatterField(m, "ingress-enabled", `{
	  "ingress": {"enabled": (if ($ARGS.named.value | type) == "boolean" then $ARGS.named.value else ($ARGS.named.value | test("^(true|1)$"; "i")) end)}
	}`, nil, nil)

	addDynamoFormatterField(m, "ingress-host", `{
	  "ingress": {"host": $ARGS.named.value}
	}`, nil, nil)

	addDynamoFormatterField(m, "ingress-path", `{
	  "ingress": {"path": $ARGS.named.value}
	}`, nil, nil)

	addDynamoFormatterField(m, "ingress-ingressClassName", `{
	  "ingress": {"ingressClassName": $ARGS.named.value}
	}`, nil, nil)

	// ModelRef fields
	addDynamoFormatterField(m, "modelRef-name", `{
	  "modelRef": {"name": $ARGS.named.value}
	}`, nil, nil)

	addDynamoFormatterField(m, "modelRef-namespace", `{
	  "modelRef": {"namespace": $ARGS.named.value}
	}`, nil, nil)

	return m
}

func addDynamoFormatterField(
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
