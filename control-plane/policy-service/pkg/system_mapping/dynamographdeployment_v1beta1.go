/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const (
	dynamoAppContainerByNameSelector  = "appContainerByName"
	dynamoInitContainerByNameSelector = "initContainerByName"
	dynamoWorkloadSpecSelector        = "workloadSpec"
	dynamoSpecPath                    = "spec"
	dynamoWorkloadMetadataSelector    = "workloadMetadata"
	dynamoMetadataPath                = "metadata"
	dynamoComponentNameArg            = "componentName"
	dynamoContainerNameArg            = "containerName"
	dynamoAllPodSpecsSelector         = "allPodSpecs"
	dynamoAllPodMetadataSelector      = "allPodMetadata"
	dynamoAllContainersSelector       = "allContainers"
	dynamoAllAppContainersSelector    = "allAppContainers"
	dynamoAllInitContainersSelector   = "allInitContainers"
	dynamoAllComponentSpecsSelector   = "allComponentSpecs"
	dynamoComponentSpecSelector       = "componentSpec"
)

func BuildSystemNvidiaComDynamoGraphDeploymentV1beta1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("nvidia.com"), Kind: "DynamoGraphDeployment", Version: "v1beta1"}, buildDynamoGraphDeploymentV1beta1SpecSelectors()
}

// Skip missing pod specs: creating an empty spec would be invalid because
// containers are required. Only pod metadata is created through dependencies.
func buildDynamoGraphDeploymentV1beta1SpecSelectors() map[string]api.SpecSelector {
	componentArgsInfo := map[string]api.MappingFieldType{
		dynamoComponentNameArg: api.MAPPINGFIELDTYPE_STRING,
	}
	containerArgsInfo := map[string]api.MappingFieldType{
		dynamoContainerNameArg: api.MAPPINGFIELDTYPE_STRING,
	}
	componentAndContainerArgsInfo := map[string]api.MappingFieldType{
		dynamoComponentNameArg: api.MAPPINGFIELDTYPE_STRING,
		dynamoContainerNameArg: api.MAPPINGFIELDTYPE_STRING,
	}

	return map[string]api.SpecSelector{
		dynamoWorkloadSpecSelector: {
			Description: convutils.Ref("the DynamoGraphDeployment spec (backendFramework and restart)"),
			Query:       `.spec`,
			FieldGroups: []string{SystemDynamoGraphDeploymentV1beta1SpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: dynamoSpecPath,
			},
		},
		dynamoWorkloadMetadataSelector: {
			Description: convutils.Ref("the DynamoGraphDeployment metadata (annotations and labels)"),
			Query:       `.metadata`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: dynamoMetadataPath,
			},
		},
		dynamoAllComponentSpecsSelector: {
			Description: convutils.Ref("all component specs in the spec.components list"),
			Query:       `.spec.components[]?`,
			FieldGroups: []string{SystemDynamoComponentV1beta1SpecFgName},
		},
		dynamoComponentSpecSelector: {
			Description: convutils.Ref("the spec of a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName)`,
			FieldGroups: []string{SystemDynamoComponentV1beta1SpecFgName},
			ArgsInfo:    componentArgsInfo,
		},
		dynamoAllPodSpecsSelector: {
			Description: convutils.Ref("pod template specs across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		dynamoAllPodMetadataSelector: {
			Description: convutils.Ref("pod template metadata across all components"),
			Query:       `.spec.components[]? | .podTemplate.metadata | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref(dynamoAllComponentSpecsSelector),
				Path:         "podTemplate.metadata",
			},
		},
		dynamoAllContainersSelector: {
			Description: convutils.Ref("app and init containers across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		dynamoAllAppContainersSelector: {
			Description: convutils.Ref("app containers across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec | .containers[]? | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		dynamoAllInitContainersSelector: {
			Description: convutils.Ref("init containers across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec | .initContainers[]? | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"allMainContainers": {
			Description: convutils.Ref("containers named main across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec | .containers[]? | select(.name == "main") | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		dynamoAppContainerByNameSelector: {
			Description: convutils.Ref("app containers selected by name across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec.containers[]? | select(.name == $ARGS.named.containerName)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    containerArgsInfo,
		},
		dynamoInitContainerByNameSelector: {
			Description: convutils.Ref("init containers selected by name across all components"),
			Query:       `.spec.components[]? | .podTemplate.spec.initContainers[]? | select(.name == $ARGS.named.containerName)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    containerArgsInfo,
		},
		"componentPodSpec": {
			Description: convutils.Ref("the pod template spec of a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentArgsInfo,
		},
		"componentPodMetadata": {
			Description: convutils.Ref("the pod template metadata of a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.metadata | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
			ArgsInfo:    componentArgsInfo,
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref(dynamoComponentSpecSelector),
				Path:         "podTemplate.metadata",
			},
		},
		"componentContainers": {
			Description: convutils.Ref("app and init containers in a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentArgsInfo,
		},
		"componentAppContainers": {
			Description: convutils.Ref("app containers in a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec | .containers[]? | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentArgsInfo,
		},
		"componentInitContainers": {
			Description: convutils.Ref("init containers in a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec | .initContainers[]? | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentArgsInfo,
		},
		"componentMainContainer": {
			Description: convutils.Ref("the container named main in a component selected by name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec | .containers[]? | select(.name == "main") | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentArgsInfo,
		},
		"componentAppContainerByName": {
			Description: convutils.Ref("an app container selected by component name and container name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec.containers[]? | select(.name == $ARGS.named.containerName)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentAndContainerArgsInfo,
		},
		"componentInitContainerByName": {
			Description: convutils.Ref("an init container selected by component name and container name"),
			Query:       `.spec.components[]? | select(.name == $ARGS.named.componentName) | .podTemplate.spec.initContainers[]? | select(.name == $ARGS.named.containerName)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
			ArgsInfo:    componentAndContainerArgsInfo,
		},
	}
}
