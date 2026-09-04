/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemRayIoRayJobV1Mapping is the system seed for ray.io/RayJob/v1.
//
// RayJob wraps a RayCluster spec under .spec.rayClusterSpec (head + workers)
// AND adds a submitter pod template under .spec.submitterPodTemplate. The
// workload-level scalars (entrypoint, runtimeEnvYAML, shutdownAfterJobFinishes,
// ttlSecondsAfterFinished, activeDeadlineSeconds, submissionMode, ...) live
// directly under .spec.
//
// NOTE: RayJob uses .spec.rayClusterSpec; RayService uses .spec.rayClusterConfig.
// Selector NAMES are kept consistent (rayClusterSpec, headPodSpec, ...) across
// kinds so users don't have to learn two vocabularies. JQ PATHS differ.
func BuildSystemRayIoRayJobV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("ray.io"), Kind: "RayJob", Version: "v1"}, buildRayIoRayJobV1SpecSelectors()
}

func buildRayIoRayJobV1SpecSelectors() map[string]api.SpecSelector {
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
			Description: convutils.Ref("the RayJob spec (entrypoint, runtimeEnvYAML, shutdownAfterJobFinishes, submissionMode, ...)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemRayJobSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},
		"workloadMetadata": {
			Description: convutils.Ref("the RayJob metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"rayClusterSpec": {
			Description: convutils.Ref("the embedded RayCluster spec under .spec.rayClusterSpec"),
			Query: `(
			  .spec.rayClusterSpec
			)`,
			FieldGroups: []string{SystemRayClusterSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "rayClusterSpec",
			},
		},

		"headPodSpec": {
			Description: convutils.Ref("the head pod template spec under .spec.rayClusterSpec.headGroupSpec.template.spec"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("rayClusterSpec"),
				Path:         "headGroupSpec.template.spec",
			},
		},
		"headPodMetadata": {
			Description: convutils.Ref("the head pod template metadata under .spec.rayClusterSpec.headGroupSpec.template.metadata"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("rayClusterSpec"),
				Path:         "headGroupSpec.template.metadata",
			},
		},

		"allWorkerPodSpecs": {
			Description: convutils.Ref("pod template specs across all worker groups under .spec.rayClusterSpec.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[].template.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"allWorkerPodMetadata": {
			Description: convutils.Ref("pod template metadata across all worker groups under .spec.rayClusterSpec.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[].template.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"workerGroupSpecByGroupName": {
			Description: convutils.Ref("a single worker group spec by groupName under .spec.rayClusterSpec.workerGroupSpecs[]"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName)
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemRayWorkerGroupSpecFgName},
		},
		"workerPodSpecByGroupName": {
			Description: convutils.Ref("the pod template spec for a single worker group by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"workerPodMetadataByGroupName": {
			Description: convutils.Ref("the pod template metadata for a single worker group by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.metadata
			)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"submitterPodSpec": {
			Description: convutils.Ref("the submitter pod template spec under .spec.submitterPodTemplate.spec"),
			Query: `(
			  .spec.submitterPodTemplate.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "submitterPodTemplate.spec",
			},
		},
		"submitterPodMetadata": {
			Description: convutils.Ref("the submitter pod template metadata under .spec.submitterPodTemplate.metadata"),
			Query: `(
			  .spec.submitterPodTemplate.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "submitterPodTemplate.metadata",
			},
		},
		"submitterContainers": {
			Description: convutils.Ref("all containers (app and init) in the submitter pod template"),
			Query: `(
			  .spec.submitterPodTemplate.spec | (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"submitterAppContainerByName": {
			Description: convutils.Ref("app container with a given name in the submitter pod template"),
			Query: `(
			  .spec.submitterPodTemplate.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"submitterInitContainerByName": {
			Description: convutils.Ref("init container with a given name in the submitter pod template"),
			Query: `(
			  (.spec.submitterPodTemplate.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodSpecs": {
			Description: convutils.Ref("pod template specs for head + all worker groups + submitter"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec,
			  .spec.rayClusterSpec.workerGroupSpecs[].template.spec,
			  .spec.submitterPodTemplate.spec
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},
		"allPodMetadata": {
			Description: convutils.Ref("pod template metadata for head + all worker groups + submitter"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.metadata,
			  .spec.rayClusterSpec.workerGroupSpecs[].template.metadata,
			  .spec.submitterPodTemplate.metadata
			) | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"headContainers": {
			Description: convutils.Ref("all containers (app and init) in the head pod template"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec | (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headAppContainers": {
			Description: convutils.Ref("app containers in the head pod template"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headInitContainers": {
			Description: convutils.Ref("init containers in the head pod template"),
			Query: `(
			  (.spec.rayClusterSpec.headGroupSpec.template.spec.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headAppContainerByName": {
			Description: convutils.Ref("app container with a given name in the head pod template"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"headInitContainerByName": {
			Description: convutils.Ref("init container with a given name in the head pod template"),
			Query: `(
			  (.spec.rayClusterSpec.headGroupSpec.template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    containerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerContainersByGroupName": {
			Description: convutils.Ref("all containers (app and init) in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerAppContainersByGroupName": {
			Description: convutils.Ref("app containers in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec.containers[]?
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerInitContainersByGroupName": {
			Description: convutils.Ref("init containers in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) |
			  (.template.spec.initContainers // [])[]
			) | select(.)`,
			ArgsInfo:    groupNameArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerAppContainerByGroupNameAndName": {
			Description: convutils.Ref("app container with a given name in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) | .template.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    groupAndContainerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"workerInitContainerByGroupNameAndName": {
			Description: convutils.Ref("init container with a given name in a worker group's pod template by groupName"),
			Query: `(
			  .spec.rayClusterSpec.workerGroupSpecs[] | select(.groupName == $ARGS.named.groupName) |
			  (.template.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo:    groupAndContainerArgsInfo,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allContainers": {
			Description: convutils.Ref("all containers (app and init) across head, all worker groups, and submitter"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec,
			  .spec.rayClusterSpec.workerGroupSpecs[].template.spec,
			  .spec.submitterPodTemplate.spec
			) | select(.) | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"allAppContainers": {
			Description: convutils.Ref("all app containers across head, all worker groups, and submitter"),
			Query: `(
			  .spec.rayClusterSpec.headGroupSpec.template.spec.containers[]?,
			  .spec.rayClusterSpec.workerGroupSpecs[].template.spec.containers[]?,
			  .spec.submitterPodTemplate.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
		"allInitContainers": {
			Description: convutils.Ref("all init containers across head, all worker groups, and submitter"),
			Query: `(
			  (.spec.rayClusterSpec.headGroupSpec.template.spec.initContainers // [])[]?,
			  (.spec.rayClusterSpec.workerGroupSpecs[].template.spec.initContainers // [])[]?,
			  (.spec.submitterPodTemplate.spec.initContainers // [])[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
