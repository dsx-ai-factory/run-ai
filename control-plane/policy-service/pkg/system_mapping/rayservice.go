/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemRayIoRayServiceV1Mapping is the system seed for ray.io/RayService/v1.
//
// RayService wraps a RayCluster spec under .spec.rayClusterConfig (head + workers)
// and exposes a Ray Serve config (serveConfigV2) + upgrade strategy at the
// workload level. There is no submitter pod (unlike RayJob).
//
// NOTE: RayService uses .spec.rayClusterConfig; RayJob uses .spec.rayClusterSpec.
// Selector NAMES are kept consistent (rayClusterSpec, headPodSpec, ...) across
// kinds so users don't have to learn two vocabularies. JQ PATHS differ.
func BuildSystemRayIoRayServiceV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("ray.io"), Kind: "RayService", Version: "v1"}, buildRayIoRayServiceV1SpecSelectors()
}

func buildRayIoRayServiceV1SpecSelectors() map[string]api.SpecSelector {
	groupNameArgsInfo := map[string]api.MappingFieldType{
		"groupName": api.MAPPINGFIELDTYPE_STRING,
	}
	groupAndContainerArgsInfo := map[string]api.MappingFieldType{
		"groupName":     api.MAPPINGFIELDTYPE_STRING,
		"containerName": api.MAPPINGFIELDTYPE_STRING,
	}
	containerArgsInfo := map[string]api.MappingFieldType{
		"containerName": api.MAPPINGFIELDTYPE_STRING,
	}

	return map[string]api.SpecSelector{
		"workloadSpec": {
			Description: convutils.Ref("the RayService spec (serveConfigV2, upgradeStrategy, ...)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemRayServiceSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},
		"workloadMetadata": {
			Description: convutils.Ref("the RayService metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"rayClusterSpec": {
			Description: convutils.Ref("the embedded RayCluster spec under .spec.rayClusterConfig"),
			Query: `(
			  .spec.rayClusterConfig
			)`,
			FieldGroups: []string{SystemRayClusterSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "rayClusterConfig",
			},
		},

		"headPodSpec": {
			Description: convutils.Ref("the head pod template spec under .spec.rayClusterConfig.headGroupSpec.template.spec"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("rayClusterSpec"),
				Path:         "headGroupSpec.template.spec",
			},
		},
		"headPodMetadata": {
			Description: convutils.Ref("the head pod template metadata under .spec.rayClusterConfig.headGroupSpec.template.metadata"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("rayClusterSpec"),
				Path:         "headGroupSpec.template.metadata",
			},
		},

		"allWorkerPodSpecs": {
			Description: convutils.Ref("pod template specs across all worker groups under .spec.rayClusterConfig.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[].template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"allWorkerPodMetadata": {
			Description: convutils.Ref("pod template metadata across all worker groups under .spec.rayClusterConfig.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[].template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"workerGroupSpecByGroupName": {
			Description: convutils.Ref("a single worker group spec by groupName under .spec.rayClusterConfig.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName)
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemRayWorkerGroupSpecFgName},
		},
		"workerPodSpecByGroupName": {
			Description: convutils.Ref("the pod template spec for a single worker group by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"workerPodMetadataByGroupName": {
			Description: convutils.Ref("the pod template metadata for a single worker group by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.metadata
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"allPodSpecs": {
			Description: convutils.Ref("pod template specs for head + all worker groups"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec,
			  .spec.rayClusterConfig.workerGroupSpecs[].template.spec
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"allPodMetadata": {
			Description: convutils.Ref("pod template metadata for head + all worker groups"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.metadata,
			  .spec.rayClusterConfig.workerGroupSpecs[].template.metadata
			) | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"headContainers": {
			Description: convutils.Ref("all containers (app and init) in the head pod template"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec | (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headAppContainers": {
			Description: convutils.Ref("app containers in the head pod template"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headInitContainers": {
			Description: convutils.Ref("init containers in the head pod template"),
			Query: `(
			  (.spec.rayClusterConfig.headGroupSpec.template.spec.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headAppContainerByName": {
			Description: convutils.Ref("app container with a given name in the head pod template"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headInitContainerByName": {
			Description: convutils.Ref("init container with a given name in the head pod template"),
			Query: `(
			  (.spec.rayClusterConfig.headGroupSpec.template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerContainersByGroupName": {
			Description: convutils.Ref("all containers (app and init) in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerAppContainersByGroupName": {
			Description: convutils.Ref("app containers in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec.containers[]?
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerInitContainersByGroupName": {
			Description: convutils.Ref("init containers in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) |
			  (.template.spec.initContainers // [])[]
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerAppContainerByGroupNameAndName": {
			Description: convutils.Ref("app container with a given name in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    groupAndContainerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerInitContainerByGroupNameAndName": {
			Description: convutils.Ref("init container with a given name in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterConfig.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) |
			  (.template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    groupAndContainerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allContainers": {
			Description: convutils.Ref("all containers (app and init) across head and all worker groups"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec,
			  .spec.rayClusterConfig.workerGroupSpecs[].template.spec
			) | select(.) | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"allAppContainers": {
			Description: convutils.Ref("all app containers across head and all worker groups"),
			Query: `(
			  .spec.rayClusterConfig.headGroupSpec.template.spec.containers[]?,
			  .spec.rayClusterConfig.workerGroupSpecs[].template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"allInitContainers": {
			Description: convutils.Ref("all init containers across head and all worker groups"),
			Query: `(
			  (.spec.rayClusterConfig.headGroupSpec.template.spec.initContainers // [])[]?,
			  (.spec.rayClusterConfig.workerGroupSpecs[].template.spec.initContainers // [])[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
