/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

const (
	SystemDynamoGraphDeploymentV1alpha1SpecFgName = "dynamoGraphDeploymentSpec"
	SystemDynamoServiceV1alpha1SpecFgName         = "dynamoServiceSpec"
)

// BuildSystemDynamoGraphDeploymentV1alpha1SpecFg returns the workload-level field group for DynamoGraphDeployment v1alpha1.
// This covers fields at .spec level (excluding services which are accessed via spec selectors).
func BuildSystemDynamoGraphDeploymentV1alpha1SpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"backendFramework": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "backendFramework",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"restart-id": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "restart.id",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"restart-strategy-type": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "restart.strategy.type",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}, map[string]api.ItemSelector{}
}

// BuildSystemDynamoServiceV1alpha1SpecFg returns the service-level field group for DynamoGraphDeployment v1alpha1 services.
// These are the "shortcut" fields at .spec.services.<name> level.
// Note: extraPodSpec is handled separately via allPodSpecs/servicePodSpec spec selectors.
func BuildSystemDynamoServiceV1alpha1SpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	return map[string]api.FieldGroupField{
		"replicas": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "replicas",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"componentType": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "componentType",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"subComponentType": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "subComponentType",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"dynamoNamespace": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "dynamoNamespace",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"globalDynamoNamespace": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "globalDynamoNamespace",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"serviceName": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "serviceName",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"sharedMemory-size": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "sharedMemory.size",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"sharedMemory-disabled": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "sharedMemory.disabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"checkpoint-enabled": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "checkpoint.enabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"checkpoint-checkpointRef": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "checkpoint.checkpointRef",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"multinode-nodeCount": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "multinode.nodeCount",
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		"ingress-enabled": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "ingress.enabled",
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		"ingress-host": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "ingress.host",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"ingress-path": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "ingress.path",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"ingress-ingressClassName": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "ingress.ingressClassName",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"modelRef-name": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "modelRef.name",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		"modelRef-namespace": {
			Formatter: SystemDynamoGraphDeploymentV1alpha1SpecFormatterName,
			SpecRef:   "modelRef.namespace",
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
	}, map[string]api.ItemSelector{}
}
