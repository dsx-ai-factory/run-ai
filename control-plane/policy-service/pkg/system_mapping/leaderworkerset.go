/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemLeaderWorkerSetXK8sIoV1Mapping is the system seed for leaderworkerset.x-k8s.io/LeaderWorkerSet/v1.
//
// LeaderWorkerSet has two pod templates (leaderTemplate, workerTemplate) under
// .spec.leaderWorkerTemplate. Each group has 1 leader and (size-1) workers.
//
// Spec selector hierarchy:
//   - workloadSpec/workloadMetadata — the LeaderWorkerSet itself (.spec, .metadata)
//   - leaderWorkerTemplate — template-level fields (size, restartPolicy, subGroupPolicy)
//   - allPodSpecs, leaderPodSpec, workerPodSpec — pod templates
//   - allContainers, leaderContainers, workerContainers — containers across/per template
func BuildSystemLeaderWorkerSetXK8sIoV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("leaderworkerset.x-k8s.io"), Kind: "LeaderWorkerSet", Version: "v1"}, buildLeaderWorkerSetXK8sIoV1SpecSelectors()
}

func buildLeaderWorkerSetXK8sIoV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		// ===== Workload Level =====
		"workloadSpec": {
			Description: convutils.Ref("the LeaderWorkerSet spec (replicas, startupPolicy, networkConfig, rolloutStrategy, leaderWorkerTemplate)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemLeaderWorkerSetSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the LeaderWorkerSet metadata (annotations and labels)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		// ===== Template Level =====
		"leaderWorkerTemplate": {
			Description: convutils.Ref("the leaderWorkerTemplate (size, restartPolicy, subGroupPolicy)"),
			Query: `(
			  .spec.leaderWorkerTemplate
			)`,
			FieldGroups: []string{SystemLeaderWorkerTemplateFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("workloadSpec"),
				Path:         "leaderWorkerTemplate",
			},
		},

		// ===== Pod Level (all templates) =====
		"allPodSpecs": {
			Description: convutils.Ref("pod specs for both leader and worker templates"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec,
			  .spec.leaderWorkerTemplate.workerTemplate.spec
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
		},

		"allPodMetadata": {
			Description: convutils.Ref("pod metadata (labels, annotations) for both leader and worker templates"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.metadata,
			  .spec.leaderWorkerTemplate.workerTemplate.metadata
			) | select(.)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		// ===== Pod Level (per template) =====
		"leaderPodSpec": {
			Description: convutils.Ref("the leader template's pod spec"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("leaderWorkerTemplate"),
				Path:         "leaderTemplate.spec",
			},
		},

		"leaderPodMetadata": {
			Description: convutils.Ref("the leader template's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("leaderWorkerTemplate"),
				Path:         "leaderTemplate.metadata",
			},
		},

		"workerPodSpec": {
			Description: convutils.Ref("the worker template's pod spec"),
			Query: `(
			  .spec.leaderWorkerTemplate.workerTemplate.spec
			)`,
			FieldGroups: []string{SystemPodSpecFgName, SystemPodSpecAndContainerFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("leaderWorkerTemplate"),
				Path:         "workerTemplate.spec",
			},
		},

		"workerPodMetadata": {
			Description: convutils.Ref("the worker template's pod metadata (labels, annotations)"),
			Query: `(
			  .spec.leaderWorkerTemplate.workerTemplate.metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				SpecSelector: convutils.Ref("leaderWorkerTemplate"),
				Path:         "workerTemplate.metadata",
			},
		},

		// ===== Container Level (all templates) =====
		"allContainers": {
			Description: convutils.Ref("all containers (app and init) across leader and worker templates"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec,
			  .spec.leaderWorkerTemplate.workerTemplate.spec
			) | select(.) | (.containers[]?, .initContainers[]?) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allAppContainers": {
			Description: convutils.Ref("all app containers across leader and worker templates"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec.containers[]?,
			  .spec.leaderWorkerTemplate.workerTemplate.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"allInitContainers": {
			Description: convutils.Ref("all init containers across leader and worker templates"),
			Query: `(
			  (.spec.leaderWorkerTemplate.leaderTemplate.spec.initContainers // [])[]?,
			  (.spec.leaderWorkerTemplate.workerTemplate.spec.initContainers // [])[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (leader template) =====
		"leaderContainers": {
			Description: convutils.Ref("all containers (app and init) in the leader template"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"leaderAppContainers": {
			Description: convutils.Ref("app containers in the leader template"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"leaderInitContainers": {
			Description: convutils.Ref("init containers in the leader template"),
			Query: `(
			  (.spec.leaderWorkerTemplate.leaderTemplate.spec.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (worker template) =====
		"workerContainers": {
			Description: convutils.Ref("all containers (app and init) in the worker template"),
			Query: `(
			  .spec.leaderWorkerTemplate.workerTemplate.spec |
			  (.containers[]?, .initContainers[]?)
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerAppContainers": {
			Description: convutils.Ref("app containers in the worker template"),
			Query: `(
			  .spec.leaderWorkerTemplate.workerTemplate.spec.containers[]?
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerInitContainers": {
			Description: convutils.Ref("init containers in the worker template"),
			Query: `(
			  (.spec.leaderWorkerTemplate.workerTemplate.spec.initContainers // [])[]
			) | select(.)`,
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		// ===== Container Level (by name) =====
		"leaderAppContainerByName": {
			Description: convutils.Ref("app container with a given name in the leader template"),
			Query: `(
			  .spec.leaderWorkerTemplate.leaderTemplate.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"leaderInitContainerByName": {
			Description: convutils.Ref("init container with a given name in the leader template"),
			Query: `(
			  (.spec.leaderWorkerTemplate.leaderTemplate.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerAppContainerByName": {
			Description: convutils.Ref("app container with a given name in the worker template"),
			Query: `(
			  .spec.leaderWorkerTemplate.workerTemplate.spec.containers[]?
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},

		"workerInitContainerByName": {
			Description: convutils.Ref("init container with a given name in the worker template"),
			Query: `(
			  (.spec.leaderWorkerTemplate.workerTemplate.spec.initContainers // [])[]
			) | select(.name == $ARGS.named.containerName)`,
			ArgsInfo: map[string]api.MappingFieldType{
				"containerName": api.MAPPINGFIELDTYPE_STRING,
			},
			FieldGroups: []string{SystemPodSpecContainerFgName, SystemPodSpecAndContainerFgName},
		},
	}
}
