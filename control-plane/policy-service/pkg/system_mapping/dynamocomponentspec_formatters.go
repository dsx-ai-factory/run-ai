/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

func BuildSystemDynamoComponentSpecFormatterFields() map[string]api.FormatterField {
	integerValue := func(minimum string) string {
		return `($ARGS.named.value | tonumber | if . >= ` + minimum + ` and . <= 2147483647 and . == floor then . else error("value must be an int32 greater than or equal to ` + minimum + `") end)`
	}
	return map[string]api.FormatterField{
		dynamoReplicasField: {
			GetExpr:  convutils.Ref(`.replicas`),
			SetExpr:  convutils.Ref(`{"replicas": ` + integerValue("0") + `}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_INTEGER},
		},
		dynamoComponentTypeField: {
			GetExpr:  convutils.Ref(`.type`),
			SetExpr:  convutils.Ref(`{"type": $ARGS.named.value}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"value": {"frontend", "worker", "prefill", "decode", "planner", "epp"}},
		},
		dynamoGlobalNamespaceField: {
			GetExpr:  convutils.Ref(`.globalDynamoNamespace`),
			SetExpr:  convutils.Ref(`{"globalDynamoNamespace": ($ARGS.named.value | if . == true or . == "true" then true elif . == false or . == "false" then false else error("value must be a boolean") end)}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_BOOLEAN},
		},
		dynamoMultinodeNodeCountField: {
			GetExpr:  convutils.Ref(`.multinode.nodeCount`),
			SetExpr:  convutils.Ref(`{"multinode": {"nodeCount": ` + integerValue("2") + `}}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_INTEGER},
		},
	}
}
