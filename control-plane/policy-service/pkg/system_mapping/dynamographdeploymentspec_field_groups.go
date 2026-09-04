/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

const (
	SystemDynamoGraphDeploymentSpecFgName = "dynamoGraphDeploymentSpec"
	SystemDynamoServiceSpecFgName         = "dynamoServiceSpec"
)

// BuildSystemDynamoGraphDeploymentSpecFg returns the workload-level field group for DynamoGraphDeployment.
// This covers fields at .spec level (excluding services which are accessed via spec selectors).
func BuildSystemDynamoGraphDeploymentSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"backendFramework": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "backendFramework",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"restart-id": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "restart.id",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"restart-strategy-type": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "restart.strategy.type",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}, map[string]api.ItemSelector{}
}

// BuildSystemDynamoServiceSpecFg returns the service-level field group for DynamoGraphDeployment services.
// These are the "shortcut" fields at .spec.services.<name> level.
// Note: extraPodSpec is handled separately via allPodSpecs/servicePodSpec spec selectors.
func BuildSystemDynamoServiceSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"replicas": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "replicas",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"componentType": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "componentType",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"subComponentType": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "subComponentType",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"dynamoNamespace": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "dynamoNamespace",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"globalDynamoNamespace": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "globalDynamoNamespace",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"serviceName": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "serviceName",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"sharedMemory-size": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "sharedMemory.size",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"sharedMemory-disabled": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "sharedMemory.disabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"checkpoint-enabled": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "checkpoint.enabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"checkpoint-checkpointRef": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "checkpoint.checkpointRef",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"multinode-nodeCount": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "multinode.nodeCount",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"ingress-enabled": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "ingress.enabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"ingress-host": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "ingress.host",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"ingress-path": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "ingress.path",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"ingress-ingressClassName": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "ingress.ingressClassName",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"modelRef-name": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "modelRef.name",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"modelRef-namespace": {
			Formatter: SystemDynamoGraphDeploymentFormatterName,
			SpecRef:   "modelRef.namespace",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}, map[string]api.ItemSelector{}
}
