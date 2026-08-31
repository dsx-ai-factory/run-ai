/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

// BuildSystemDeploymentSpecFormatterFields builds the runai-deploymentSpec system formatter,
// holding fields whose argsEnum we want enforced at write time. The setExprs are rooted at
// the spec-selector anchor (the workload .spec).
func BuildSystemDeploymentSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// strategy-Type
	////////////////////////////////////////////////////////
	addDeploymentSpecField(m, "strategy-Type", `{
	  "strategy": {
		"type": $ARGS.named.value
	  }
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"Recreate", "RollingUpdate"},
		})

	return m
}

func addDeploymentSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr:  &setExpr,
		GetExpr:  getExpr,
		ArrayKey: arrayKey,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
