/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
)

const (
	SystemDynamoGraphDeploymentV1beta1SpecFgName = "dynamoGraphDeploymentV1beta1Spec"
	dynamoBackendFrameworkField                  = "backendFramework"
	dynamoRestartIDField                         = "restart-id"
	dynamoRestartStrategyTypeField               = "restart-strategy-type"
)

func BuildSystemDynamoGraphDeploymentV1beta1SpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return BuildFieldGroupFieldMap(dynamoGraphDeploymentV1beta1SpecFgFields()), nil
}

func dynamoGraphDeploymentV1beta1SpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	formatter := SystemDynamoGraphDeploymentV1beta1SpecFormatterName
	return []SystemFgField{
		{
			Name:      dynamoBackendFrameworkField,
			Formatter: formatter,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      dynamoRestartIDField,
			Formatter: keyValue,
			SpecRef:   strRef("restart.id"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      dynamoRestartStrategyTypeField,
			Formatter: formatter,
			SpecRef:   strRef("restart.strategy.type"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
