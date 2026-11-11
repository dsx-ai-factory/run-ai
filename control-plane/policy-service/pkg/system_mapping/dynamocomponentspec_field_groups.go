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
	SystemDynamoComponentSpecFgName = "dynamoComponentSpec"
	dynamoReplicasField             = "replicas"
	dynamoComponentTypeField        = "type"
	dynamoGlobalNamespaceField      = "globalDynamoNamespace"
	dynamoMultinodeNodeCountField   = "multinode-nodeCount"
)

func BuildSystemDynamoComponentSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return BuildFieldGroupFieldMap(dynamoComponentSpecFgFields()), nil
}

func dynamoComponentSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	formatter := SystemDynamoComponentSpecFormatterName
	return []SystemFgField{
		{
			Name:      dynamoReplicasField,
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      dynamoComponentTypeField,
			Formatter: formatter,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      dynamoGlobalNamespaceField,
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      dynamoMultinodeNodeCountField,
			Formatter: keyValue,
			SpecRef:   strRef("multinode.nodeCount"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
	}
}
