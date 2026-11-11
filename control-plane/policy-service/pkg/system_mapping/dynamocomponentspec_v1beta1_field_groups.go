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
	SystemDynamoComponentV1beta1SpecFgName        = "dynamoComponentV1beta1Spec"
	dynamoComponentV1beta1ReplicasField           = "replicas"
	dynamoComponentV1beta1TypeField               = "type"
	dynamoComponentV1beta1GlobalNamespaceField    = "globalDynamoNamespace"
	dynamoComponentV1beta1MultinodeNodeCountField = "multinode-nodeCount"
	dynamoComponentV1beta1SharedMemorySizeField   = "sharedMemorySize"
	dynamoComponentV1beta1ModelRefNameField       = "modelRef-name"
)

func BuildSystemDynamoComponentV1beta1SpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return BuildFieldGroupFieldMap(dynamoComponentV1beta1SpecFgFields()), nil
}

func dynamoComponentV1beta1SpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	formatter := SystemDynamoComponentV1beta1SpecFormatterName
	return []SystemFgField{
		{
			Name:      dynamoComponentV1beta1ReplicasField,
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      dynamoComponentV1beta1TypeField,
			Formatter: formatter,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      dynamoComponentV1beta1GlobalNamespaceField,
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      dynamoComponentV1beta1MultinodeNodeCountField,
			Formatter: keyValue,
			SpecRef:   strRef("multinode.nodeCount"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      dynamoComponentV1beta1SharedMemorySizeField,
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      dynamoComponentV1beta1ModelRefNameField,
			Formatter: keyValue,
			SpecRef:   strRef("modelRef.name"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}
}
