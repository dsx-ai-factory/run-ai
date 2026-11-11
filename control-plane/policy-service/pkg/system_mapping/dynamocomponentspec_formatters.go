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
	return map[string]api.FormatterField{
		dynamoComponentTypeField: {
			GetExpr:  convutils.Ref(`.type`),
			SetExpr:  convutils.Ref(`{"type": $ARGS.named.value}`),
			ArgsInfo: map[string]api.MappingFieldType{"value": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"value": {"frontend", "worker", "prefill", "decode", "planner", "epp"}},
		},
	}
}
